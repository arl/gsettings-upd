# gsettings-upd: A systemd service to update gsettings values.

Nearly everything in Gnome is configurable via gsettings. You can list all the
values managed by gsettings with:

```
gsettings list-recursively
```

For example, you want to change desktop wallpapers every hour, ok:

```json
{
  "actions": [
    {
      "schema": "org.gnome.desktop.background",
      "key": "picture-uri",
      "values": [
        "file:///path/to/my/wallpapers/image1.png",
        "file:///path/to/my/wallpapers/image2.png",
        "file:///path/to/my/wallpapers/image3.png"
      ],
      "every": "1h"
    }
  ]
}

```
