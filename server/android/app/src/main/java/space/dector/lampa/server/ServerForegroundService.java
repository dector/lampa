package space.dector.lampa.server;

import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.app.PendingIntent;
import android.app.Service;
import android.content.Context;
import android.content.Intent;
import android.content.SharedPreferences;
import android.os.Build;
import android.os.IBinder;
import android.util.Log;

import androidx.annotation.Nullable;
import androidx.core.app.NotificationCompat;

import exported.Server;
import exported.ServerConfig;

public class ServerForegroundService extends Service {
    public static final String ACTION_START_SERVER = "space.dector.lampa.server.action.START_SERVER";
    public static final String ACTION_STOP_SERVER = "space.dector.lampa.server.action.STOP_SERVER";
    public static final String ACTION_TOGGLE_SERVER = "space.dector.lampa.server.action.TOGGLE_SERVER";
    public static final String ACTION_SERVER_STATE_CHANGED = "space.dector.lampa.server.action.SERVER_STATE_CHANGED";
    public static final String EXTRA_IS_RUNNING = "space.dector.lampa.server.extra.IS_RUNNING";
    public static final String EXTRA_PROXY_ENDPOINT = "space.dector.lampa.server.extra.PROXY_ENDPOINT";
    public static final String EXTRA_CONTROL_ENDPOINT = "space.dector.lampa.server.extra.CONTROL_ENDPOINT";

    private static final String CHANNEL_ID = "server_control_channel_v2";
    private static final int NOTIFICATION_ID = 1001;
    private static final String TAG = "ServerForegroundService";
    private static final String PREFS_SERVER_STATE = "space.dector.lampa.server.prefs.SERVER_STATE";
    private static final String PREF_IS_RUNNING = "is_running";
    private static final String PREF_PROXY_ENDPOINT = "proxy_endpoint";
    private static final String PREF_CONTROL_ENDPOINT = "control_endpoint";

    private final ServerConfig serverConfig = ServerRuntimeConfig.toServerConfig();
    private final Server server = new Server(serverConfig);

    @Override
    public void onCreate() {
        super.onCreate();
        createNotificationChannel();
        startForeground(NOTIFICATION_ID, buildNotification(server.isRunning(), null));
        notifyServerStateChanged();
    }

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        String action = intent != null ? intent.getAction() : null;

        if (ACTION_START_SERVER.equals(action)) {
            startServer();
        } else if (ACTION_STOP_SERVER.equals(action)) {
            stopServer();
        } else if (ACTION_TOGGLE_SERVER.equals(action)) {
            if (server.isRunning()) {
                stopServer();
            } else {
                startServer();
            }
        }

        updateNotification(null);
        notifyServerStateChanged();
        return START_STICKY;
    }

    private void startServer() {
        if (server.isRunning()) {
            return;
        }

        String result = server.startAsync();
        if (!result.isEmpty()) {
            Log.e(TAG, "Failed to start server: " + result
                    + " | proxy=" + serverConfig.getProxyListenAddress()
                    + " control=" + serverConfig.getControlListenAddress());
            updateNotification("Start failed");
        }
    }

    private void stopServer() {
        if (!server.isRunning()) {
            return;
        }

        String result = server.stop();
        if (!result.isEmpty()) {
            Log.e(TAG, "Failed to stop server: " + result);
            updateNotification("Stop failed");
        }
    }

    private void updateNotification(@Nullable String extraMessage) {
        Notification notification = buildNotification(server.isRunning(), extraMessage);
        NotificationManager manager = getSystemService(NotificationManager.class);
        manager.notify(NOTIFICATION_ID, notification);
    }

    private void notifyServerStateChanged() {
        boolean isRunning = server.isRunning();
        String proxyEndpoint = serverConfig.getProxyListenAddress();
        String controlEndpoint = serverConfig.getControlListenAddress();

        persistState(this, isRunning, proxyEndpoint, controlEndpoint);

        Intent stateIntent = new Intent(ACTION_SERVER_STATE_CHANGED)
                .setPackage(getPackageName())
                .putExtra(EXTRA_IS_RUNNING, isRunning)
                .putExtra(EXTRA_PROXY_ENDPOINT, proxyEndpoint)
                .putExtra(EXTRA_CONTROL_ENDPOINT, controlEndpoint);
        sendBroadcast(stateIntent);
    }

    public static boolean lastKnownIsRunning(Context context) {
        SharedPreferences prefs = context.getSharedPreferences(PREFS_SERVER_STATE, Context.MODE_PRIVATE);
        return prefs.getBoolean(PREF_IS_RUNNING, false);
    }

    public static String lastKnownProxyEndpoint(Context context) {
        SharedPreferences prefs = context.getSharedPreferences(PREFS_SERVER_STATE, Context.MODE_PRIVATE);
        return prefs.getString(PREF_PROXY_ENDPOINT, ServerRuntimeConfig.proxyEndpoint());
    }

    public static String lastKnownControlEndpoint(Context context) {
        SharedPreferences prefs = context.getSharedPreferences(PREFS_SERVER_STATE, Context.MODE_PRIVATE);
        return prefs.getString(PREF_CONTROL_ENDPOINT, ServerRuntimeConfig.controlEndpoint());
    }

    private static void persistState(Context context, boolean isRunning, String proxyEndpoint, String controlEndpoint) {
        SharedPreferences prefs = context.getSharedPreferences(PREFS_SERVER_STATE, Context.MODE_PRIVATE);
        prefs.edit()
                .putBoolean(PREF_IS_RUNNING, isRunning)
                .putString(PREF_PROXY_ENDPOINT, proxyEndpoint)
                .putString(PREF_CONTROL_ENDPOINT, controlEndpoint)
                .apply();
    }

    private Notification buildNotification(boolean isRunning, @Nullable String extraMessage) {
        String title = getString(R.string.notification_title);
        String stateText = isRunning
                ? getString(R.string.notification_state_running)
                : getString(R.string.notification_state_stopped);

        String contentText = extraMessage == null ? stateText : stateText + " • " + extraMessage;

        Intent toggleIntent = new Intent(this, ServerForegroundService.class);
        toggleIntent.setAction(ACTION_TOGGLE_SERVER);

        PendingIntent togglePendingIntent = PendingIntent.getService(
                this,
                0,
                toggleIntent,
                PendingIntent.FLAG_UPDATE_CURRENT | PendingIntent.FLAG_IMMUTABLE
        );

        Intent appIntent = new Intent(this, MainActivity.class);
        PendingIntent appPendingIntent = PendingIntent.getActivity(
                this,
                0,
                appIntent,
                PendingIntent.FLAG_UPDATE_CURRENT | PendingIntent.FLAG_IMMUTABLE
        );

        String actionLabel = isRunning
                ? getString(R.string.notification_action_turn_off)
                : getString(R.string.notification_action_turn_on);

        return new NotificationCompat.Builder(this, CHANNEL_ID)
                .setSmallIcon(android.R.drawable.stat_notify_sync)
                .setContentTitle(title)
                .setContentText(contentText)
                .setOngoing(true)
                .setVisibility(NotificationCompat.VISIBILITY_PUBLIC)
                .setCategory(NotificationCompat.CATEGORY_SERVICE)
                .setPriority(NotificationCompat.PRIORITY_HIGH)
                .setForegroundServiceBehavior(NotificationCompat.FOREGROUND_SERVICE_IMMEDIATE)
                .setContentIntent(appPendingIntent)
                .addAction(0, actionLabel, togglePendingIntent)
                .build();
    }

    private void createNotificationChannel() {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.O) {
            return;
        }

        NotificationChannel channel = new NotificationChannel(
                CHANNEL_ID,
                getString(R.string.notification_channel_name),
                NotificationManager.IMPORTANCE_HIGH
        );
        channel.setDescription(getString(R.string.notification_channel_description));
        channel.setLockscreenVisibility(Notification.VISIBILITY_PUBLIC);

        NotificationManager manager = getSystemService(NotificationManager.class);
        manager.createNotificationChannel(channel);
    }

    @Nullable
    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }
}
