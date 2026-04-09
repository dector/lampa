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
