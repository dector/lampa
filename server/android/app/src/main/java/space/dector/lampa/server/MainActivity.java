package space.dector.lampa.server;

import android.os.Bundle;
import android.widget.CompoundButton;
import android.widget.Toast;
import android.widget.ToggleButton;

import androidx.appcompat.app.AppCompatActivity;

import exported.Server;

public class MainActivity extends AppCompatActivity {
    private Server server;
    private ToggleButton toggleButton;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);

        server = new Server();
        toggleButton = findViewById(R.id.serverToggleButton);

        // Restore UI state if app process was recreated while server is still running.
        toggleButton.setChecked(server.isRunning());

        toggleButton.setOnCheckedChangeListener(this::onToggleChanged);
    }

    private void onToggleChanged(CompoundButton buttonView, boolean isChecked) {
        String result;

        if (isChecked) {
            result = server.startAsync();
            if (result.isEmpty()) {
                Toast.makeText(this, "Server ON", Toast.LENGTH_SHORT).show();
            } else {
                buttonView.setChecked(false);
                Toast.makeText(this, "Failed to start: " + result, Toast.LENGTH_LONG).show();
            }
        } else {
            result = server.stop();
            if (result.isEmpty()) {
                Toast.makeText(this, "Server OFF", Toast.LENGTH_SHORT).show();
            } else {
                buttonView.setChecked(true);
                Toast.makeText(this, "Failed to stop: " + result, Toast.LENGTH_LONG).show();
            }
        }
    }

    @Override
    protected void onDestroy() {
        super.onDestroy();
        // Keep behavior simple: when app closes, stop server.
        server.stop();
    }
}
