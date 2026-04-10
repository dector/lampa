package space.dector.lampa.server;

import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.app.PendingIntent;
import android.app.Service;
import android.content.Intent;
import android.os.Build;
import android.os.IBinder;
import android.util.Log;

import androidx.annotation.Nullable;
import androidx.core.app.NotificationCompat;

import exported.Server;

public class ServerForegroundService extends Service {
    public static final String ACTION_START_SERVER = "space.dector.lampa.server.action.START_SERVER";
    public static final String ACTION_STOP_SERVER = "space.dector.lampa.server.action.STOP_SERVER";
    public static final String ACTION_TOGGLE_SERVER = "space.dector.lampa.server.action.TOGGLE_SERVER";
    public static final String ACTION_SERVER_STATE_CHANGED = "space.dector.lampa.server.action.SERVER_STATE_CHANGED";
    public static final String EXTRA_IS_RUNNING = "space.dector.lampa.server.extra.IS_RUNNING";

    private static final String CHANNEL_ID = "server_control_channel_v2";
    private static final int NOTIFICATION_ID = 1001;
    private static final String TAG = "ServerForegroundService";

    private final Server server = new Server();

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
            Log.e(TAG, "Failed to start server: " + result);
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
        Intent stateIntent = new Intent(ACTION_SERVER_STATE_CHANGED)
                .setPackage(getPackageName())
                .putExtra(EXTRA_IS_RUNNING, server.isRunning());
        sendBroadcast(stateIntent);
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
