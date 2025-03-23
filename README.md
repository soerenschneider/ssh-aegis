# SSH-Aegis

🚀 **SSH-Aegis** is a lightweight, zero-dependency tool that dynamically adjusts your SSH listening address based on the status of your wireguard VPN connection. It helps **reduce SSH attack surface** by ensuring that SSH is only exposed publicly when absolutely necessary.

## ✨ Features
- **Automatic SSH Binding:**
  - If VPN is **UP**, SSH listens only on a predefined internal VPN IP.
  - If VPN is **DOWN**, SSH falls back to listening on a **public address**.
- **Zero Dependencies:** Built using only the Go standard library.
- **Configurable via JSON:** Define VPN interface, IPs, .
- **Metrics Support:** Provides real-time status monitoring.

## 🔐 Security Benefits
✅ **Minimizes SSH Exposure** → Attackers cannot reach SSH when VPN is active.  
✅ **Automated Failover** → Ensures access if VPN disconnects.  
✅ **Zero Dependencies** → No external libraries, reducing supply chain risks.

## 🛠️ How It Works
SSH-Aegis continuously monitors the VPN interface:
1. **VPN Active → SSH on Internal IP** 🛡️
   - Reduces exposure to potential SSH 0-days.
   - Only accessible from within the VPN.
2. **VPN Down → SSH on Public IP** ⚠️
   - Ensures remote access remains available.
   - Auto-switches without manual intervention.

## 📄 Configuration
Create a JSON configuration file (e.g., `config.json`):

```json
{
  "up": ["10.8.0.1"],
  "down": ["0.0.0.0"],
  "sshd_config_file": "/etc/ssh/sshd_config",
  "ssh_service_name": "ssh",
  "metrics_file": "/var/lib/node_exporter/ssh-aegis.prom",
  "wg": "wg0"
}
```

### Parameters:

## 📝 Explanation of Fields

| Key                    | Type       | Description                                                                 | Default                               | Optional |
|------------------------|------------|-----------------------------------------------------------------------------|---------------------------------------|----------|
| **`up`**               | `[]string` | Addresses to set when the VPN is **UP** (e.g., internal VPN IP).            |                                       |          |
| **`down`**             | `[]string` | Addresses to set when the VPN is **DOWN** (e.g., public IP).                | 0.0.0.0                               |          |
| **`unknown`**          | `[]string` | Addresses to set when VPN status is **unknown** (e.g., temporary failover). |                                       |          |
| **`sshd_config_file`** | `string`   | Path to the SSHD configuration file.                                        | /etc/ssh/sshd_config                  |          |
| **`wg`**               | `string`   | Name of the WireGuard interface being monitored.                            | wg0                                   |          |
| **`ssh_service_name`** | `string`   | Name of the SSH service to restart.                                         | sshd                                  |          |
| **`metrics_file`**     | `string`   | Path to a file where SSH-Aegis logs metrics.                                | /var/lib/node_exporter/ssh_aegis.prom |          |

## 🚀 Usage
Run SSH-Aegis as a background service:

```sh
# ssh-aegis --help
Usage of ssh-aegis:
  -config string
        Path of config file (default "/etc/ssh-aegis.json")
  -debug
        Print debug logs
  -version
        Print version and exit


# ssh-aegis -config config.json
```

## 📊 Metrics & Monitoring
SSH-Aegis exposes metrics on via Prometheus NodeExporter.


| Metric Name                                          | Type      | Description                                                            |
|------------------------------------------------------|-----------|------------------------------------------------------------------------|
| **`ssh_aegis_timestamp_seconds`**                    | `gauge`   | The timestamp of the last SSH-Aegis invocation.                        |
| **`ssh_aegis_status`**                               | `gauge`   | Represents the current VPN tunnel status (`up`, `down`, or `unknown`). |
| **`ssh_aegis_last_status_change_timestamp_seconds`** | `gauge`   | Timestamp of the last VPN status change.                               |
| **`ssh_aegis_restart_ssh_errors`**                   | `counter` | Number of errors encountered while restarting the SSH service.         |
| **`ssh_aegis_read_config_errors`**                   | `counter` | Number of errors encountered while reading the configuration file.     |
| **`ssh_aegis_write_config_errors`**                  | `counter` | Number of errors encountered while writing to the configuration file.  |



### Systemd Service (Linux)
To ensure SSH-Aegis runs on startup, create a systemd service:

```ini
[Unit]
Description=SSH-Aegis - Dynamic SSH Listener
After=network.target

[Service]
ExecStart=/usr/local/bin/ssh-aegis
Restart=always

[Install]
WantedBy=multi-user.target
```

