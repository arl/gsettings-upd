package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-systemd/daemon"
)

var (
	errNoValues = errors.New("action with no values")
	errNoSchema = errors.New("action with empty schema")
	errNoKey    = errors.New("action with empty key")
)

const (
	// cfgEnvVar contains the name of the environment variable pointing to the
	// configuration file.
	cfgEnvVar = "GSETTINGS_UPD_CONFIG"
)

// sets the value of a GSetting key
func gset(schema, key, val string) error {
	cmd := exec.Command("/usr/bin/gsettings", "set", schema, key, val)
	return cmd.Run()
}

// gets the value of a GSetting key
func gget(schema, key string) (string, error) {
	cmd := exec.Command("/usr/bin/gsettings", "get", schema, key)
	b, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.Trim(strings.TrimSpace(string(b)), "'"), nil
}

func notify(state string) error {
	notified, err := daemon.SdNotify(false, state)
	if notified {
		return nil
	}
	if err == nil {
		return fmt.Errorf("notification not supported")
	}
	return err
}

type stopTickFunc func()

// setupWatchdog sets up systemd watchdog notification. It returns a channel
// 'ticking' at the interval suitable for the systemd watchdog. In case watchdog
// has been disabled for the service, it returns a nil channel (blocking).
func setupWatchdog() (<-chan time.Time, stopTickFunc) {
	µsec := os.Getenv("WATCHDOG_USEC")
	i, err := strconv.ParseInt(µsec, 10, 64)
	if err != nil {
		log.Println("can't parse WATCHDOG_USEC:", err)
		return nil, func() {}
	}

	// use a `WATCHDOG_USEC / 2` watchdog notification period
	t := time.NewTicker(time.Duration(i) * time.Microsecond / 2)

	return t.C, func() { t.Stop() }
}

type config struct {
	Actions []Action
}

func loadConfig(b []byte) (*config, error) {
	var cfg config
	err := json.Unmarshal(b, &cfg)

	// check config validity
	for _, a := range cfg.Actions {
		if len(a.Values) == 0 {
			return nil, errNoValues
		}
		if a.Schema == "" {
			return nil, errNoSchema
		}
		if a.Key == "" {
			return nil, errNoKey
		}
	}
	return &cfg, err
}

func startTicker(done <-chan struct{}, a *Action) {
	if len(a.Values) == 0 {
		log.Printf("no values to choose from for %v/%v", a.Schema, a.Key)
		return
	}

	ticker := time.NewTicker(a.Every.Duration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := a.perform()
			if err != nil {
				log.Printf("can't update value (schema=%v, key=%v): %v", a.Schema, a.Key, err)
				return
			} else {
				log.Printf("successfully updated value (schema=%v, key=%v)", a.Schema, a.Key)
			}
		case <-done:
			return
		}
	}

	log.Printf("leaving startTicker %v %v", a.Schema, a.Key)
}

func main() {

	fmt.Println("os.Getuid: ", os.Getuid())

	// obtain configuration file
	ev := os.Getenv(cfgEnvVar)
	buf, err := ioutil.ReadFile(ev)
	if err != nil {
		log.Fatalf("can't load configuration: %v", err)
	}

	var cfg *config
	cfg, err = loadConfig(buf)
	if err != nil {
		log.Fatalf("can't parse configuration: %v", err)
	}
	fmt.Println("loaded config", cfg)

	// Set up a done channel that will be passed to all goroutines,
	// to use as an exit signal.
	done := make(chan struct{})
	defer close(done)

	var wg sync.WaitGroup
	wg.Add(len(cfg.Actions))
	for i := range cfg.Actions {
		go func(i int) {
			defer wg.Done()
			startTicker(done, &cfg.Actions[i])
		}(i)
	}

	watchdog, stop := setupWatchdog()
	defer stop()

	wg.Add(1)
	go func() {
		for {
			defer wg.Done()
			select {
			case <-watchdog:
				err := notify("WATCHDOG=1")
				if err != nil {
					log.Println("can't update watchdog timestamp:", err)
				}
			case <-done:
				return
			}
		}
	}()

	// inform systemd we are ready
	err = notify("READY=1")
	if err != nil {
		log.Fatalln("can't notify readiness:", err)
	}
	wg.Wait()
	log.Println("gsettings-upd quitting")
}
