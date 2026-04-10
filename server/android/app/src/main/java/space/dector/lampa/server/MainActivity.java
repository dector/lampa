package space.dector.lampa.server;

import android.Manifest;
import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.content.IntentFilter;
import android.content.pm.PackageManager;
import android.net.Uri;
import android.os.Build;
import android.os.Bundle;
import android.provider.Settings;
import android.view.View;
import android.widget.CompoundButton;
import android.widget.TextView;
import android.widget.ToggleButton;

import androidx.appcompat.app.AppCompatActivity;
import androidx.core.app.NotificationManagerCompat;
import androidx.core.content.ContextCompat;

import exported.Server;

public class MainActivity extends AppCompatActivity {
    private ToggleButton toggleButton;
    private TextView notificationPermissionBanner;
    private boolean suppressToggleCallback;

    private final BroadcastReceiver serverStateReceiver = new BroadcastReceiver() {
        @Override
        public void onReceive(Context context, Intent intent) {
            if (!ServerForegroundService.ACTION_SERVER_STATE_CHANGED.equals(intent.getAction())) {
                return;
            }

            boolean isRunning = intent.getBooleanExtra(ServerForegroundService.EXTRA_IS_RUNNING, false);
            setToggleCheckedSafely(isRunning);
        }
    };

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);

        toggleButton = findViewById(R.id.serverToggleButton);
        notificationPermissionBanner = findViewById(R.id.notificationPermissionBanner);

        toggleButton.setOnCheckedChangeListener(this::onToggleChanged);
        notificationPermissionBanner.setOnClickListener(v -> openNotificationSettings());

        syncToggleWithServerState();
    }

    @Override
    protected void onStart() {
        super.onStart();
        IntentFilter intentFilter = new IntentFilter(ServerForegroundService.ACTION_SERVER_STATE_CHANGED);
        ContextCompat.registerReceiver(this, serverStateReceiver, intentFilter, ContextCompat.RECEIVER_NOT_EXPORTED);
    }

    @Override
    protected void onStop() {
        super.onStop();
        unregisterReceiver(serverStateReceiver);
    }

    @Override
    protected void onResume() {
        super.onResume();
        syncToggleWithServerState();
        updateNotificationPermissionBanner();
    }

    private void updateNotificationPermissionBanner() {
        notificationPermissionBanner.setVisibility(hasNotificationPermission() ? View.GONE : View.VISIBLE);
    }

    private boolean hasNotificationPermission() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU
                && checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS) != PackageManager.PERMISSION_GRANTED) {
            return false;
        }

        return NotificationManagerCompat.from(this).areNotificationsEnabled();
    }

    private void openNotificationSettings() {
        Intent intent = new Intent(Settings.ACTION_APP_NOTIFICATION_SETTINGS)
                .putExtra(Settings.EXTRA_APP_PACKAGE, getPackageName());

        if (intent.resolveActivity(getPackageManager()) != null) {
            startActivity(intent);
            return;
        }

        Intent fallbackIntent = new Intent(
                Settings.ACTION_APPLICATION_DETAILS_SETTINGS,
                Uri.fromParts("package", getPackageName(), null)
        );
        startActivity(fallbackIntent);
    }

    private void syncToggleWithServerState() {
        setToggleCheckedSafely(new Server().isRunning());
    }

    private void setToggleCheckedSafely(boolean checked) {
        suppressToggleCallback = true;
        toggleButton.setChecked(checked);
        suppressToggleCallback = false;
    }

    private void onToggleChanged(CompoundButton buttonView, boolean isChecked) {
        if (suppressToggleCallback) {
            return;
        }

        String action = isChecked
                ? ServerForegroundService.ACTION_START_SERVER
                : ServerForegroundService.ACTION_STOP_SERVER;

        Intent serviceIntent = new Intent(this, ServerForegroundService.class);
        serviceIntent.setAction(action);

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            startForegroundService(serviceIntent);
        } else {
            startService(serviceIntent);
        }
    }
}
