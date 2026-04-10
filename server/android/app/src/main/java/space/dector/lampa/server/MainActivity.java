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


public class MainActivity extends AppCompatActivity {
    private ToggleButton toggleButton;
    private TextView notificationPermissionBanner;
    private TextView proxyEndpointText;
    private TextView controlEndpointText;
    private boolean suppressToggleCallback;

    private final BroadcastReceiver serverStateReceiver = new BroadcastReceiver() {
        @Override
        public void onReceive(Context context, Intent intent) {
            if (!ServerForegroundService.ACTION_SERVER_STATE_CHANGED.equals(intent.getAction())) {
                return;
            }

            boolean isRunning = intent.getBooleanExtra(ServerForegroundService.EXTRA_IS_RUNNING, false);
            String proxyEndpoint = intent.getStringExtra(ServerForegroundService.EXTRA_PROXY_ENDPOINT);
            String controlEndpoint = intent.getStringExtra(ServerForegroundService.EXTRA_CONTROL_ENDPOINT);

            setToggleCheckedSafely(isRunning);
            bindEndpoints(proxyEndpoint, controlEndpoint);
        }
    };

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);

        toggleButton = findViewById(R.id.serverToggleButton);
        notificationPermissionBanner = findViewById(R.id.notificationPermissionBanner);
        proxyEndpointText = findViewById(R.id.proxyEndpointText);
        controlEndpointText = findViewById(R.id.controlEndpointText);

        toggleButton.setOnCheckedChangeListener(this::onToggleChanged);
        notificationPermissionBanner.setOnClickListener(v -> openNotificationSettings());

        bindEndpoints(
                ServerForegroundService.lastKnownProxyEndpoint(this),
                ServerForegroundService.lastKnownControlEndpoint(this)
        );
        setToggleCheckedSafely(ServerForegroundService.lastKnownIsRunning(this));
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

    private void bindEndpoints(String proxyEndpoint, String controlEndpoint) {
        String proxy = (proxyEndpoint == null || proxyEndpoint.isEmpty())
                ? ServerRuntimeConfig.proxyEndpoint()
                : proxyEndpoint;
        String control = (controlEndpoint == null || controlEndpoint.isEmpty())
                ? ServerRuntimeConfig.controlEndpoint()
                : controlEndpoint;

        proxyEndpointText.setText(getString(R.string.proxy_endpoint_value, proxy));
        controlEndpointText.setText(getString(R.string.control_endpoint_value, control));
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
