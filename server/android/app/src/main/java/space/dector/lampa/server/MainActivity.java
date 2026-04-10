package space.dector.lampa.server;

import android.Manifest;
import android.content.Intent;
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

import exported.Server;

public class MainActivity extends AppCompatActivity {
    private ToggleButton toggleButton;
    private TextView notificationPermissionBanner;
    private boolean suppressToggleCallback;

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
        suppressToggleCallback = true;
        toggleButton.setChecked(new Server().isRunning());
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
