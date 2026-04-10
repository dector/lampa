# Lampa Android Server Toggle App

Small Java Android app with one toggle button to control the Go server from `lampa-server.aar`.

## AAR dependency (relative path)

The app module uses:

```gradle
implementation files('../../core/build/android/lampa-server.aar')
```

So the expected AAR location is:

- `server/core/build/android/lampa-server.aar`

## Build the AAR first

From repository root:

```bash
./server/core/scripts/build-aar.sh
```

## Open/build app

Open `server/android` in Android Studio and run the `app` module.

## Forward Android control port to host

After the app is running on a connected device/emulator, forward the control port to your machine:

```bash
adb forward tcp:46899 tcp:46899
```

If you use a custom control port, replace `46899` with your port value.

## Verify with `lampa server ping`

From repository root (or any shell with `lampa` available), verify connectivity:

```bash
lampa server ping --port 46899
```

You can omit `--port` when using the default port.
