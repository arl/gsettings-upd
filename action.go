package main

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"time"
)

// Duration is a time.Duration that can be unmarshalled to JSON.
type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("invalid duration")
}

type Action struct {
	Schema string     // Schema is the targeted gsettings schema.
	Key    string     // Key is the key modified.
	Every  Duration   // Every is the period at which the key is updated.
	Random bool       // Random indicates wether the next is randomly chosen.
	Values []string   // Values is the set of values from which next value if chosen.
	rng    *rand.Rand // rng will be lazily created if Random is true.
}

// perform performs the action.
func (a *Action) perform() error {
	v := a.nextValue()
	return gset(a.Schema, a.Key, v)
}

// nextValue returns the value to which we should set the key corresponding to current
// action.
func (a *Action) nextValue() string {
	var nexti int
	if a.Random {
		if a.rng == nil {
			// lazily create the random number generator
			seed := time.Now().UnixNano()
			a.rng = rand.New(rand.NewSource(seed))
		}
		nexti = a.rng.Intn(len(a.Values))
	} else {
		cur, err := gget(a.Schema, a.Key)
		if err != nil {
			log.Printf("can't get value (schema=%v, key=%v): %v", a.Schema, a.Key, err)
		} else {
			for i := range a.Values {
				if a.Values[i] == cur {
					nexti = (i + 1) % len(a.Values)
					break
				}
			}
		}
	}
	return a.Values[nexti]
}
