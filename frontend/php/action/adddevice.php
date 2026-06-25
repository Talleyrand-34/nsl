
<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch select options
$models = json_decode(@file_get_contents(MODELS_ENDPOINT), true) ?: [];
$zones = json_decode(@file_get_contents(ZONES_ENDPOINT), true) ?: [];
$proprietaries = json_decode(@file_get_contents(PROPRIETARIES_ENDPOINT), true) ?: [];
$scanProfiles = json_decode(@file_get_contents(SCAN_PROFILES_ENDPOINT), true) ?: [];

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $label = trim($_POST['label'] ?? '');
    $model = $_POST['model'] ?? '';
    $zoneid = $_POST['zoneid'] ?? '';
    $proprietary = $_POST['proprietary'] ?? '';
    $profile = $_POST['profile'] ?? '';

    // Collect IPs from dynamic inputs
    $ips = [];
    if (isset($_POST['ips']) && is_array($_POST['ips'])) {
        foreach ($_POST['ips'] as $ip) {
            $ip = trim($ip);
            if (!empty($ip)) {
                $ips[] = $ip;
            }
        }
    }

    // Lookup zone name by id
    $zonename = '';
    if ($zoneid !== '') {
        foreach ($zones as $zone) {
            if ((string) $zone['id'] === $zoneid) {
                $zonename = $zone['name'];
                break;
            }
        }
    }

    if ($label === '') {
        $message = 'Please enter a device label.';
    } elseif ($model === '') {
        $message = 'Please select a model.';
    } elseif ($zoneid === '') {
        $message = 'Please select a zone.';
    } elseif ($proprietary === '') {
        $message = 'Please select a proprietary.';
    } else {
        $data = json_encode([
            'label' => $label,
            'model_name' => $model,
            'zone_id' => $zoneid,
            'zone_name' => $zonename,
            'proprietary' => $proprietary,
            'profile' => $profile,
            'ips' => $ips
        ]);

        $ch = curl_init(DEVICES_ENDPOINT);
        curl_setopt($ch, CURLOPT_CUSTOMREQUEST, 'POST');
        curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'Content-Type: application/json',
            'Content-Length: ' . strlen($data)
        ]);

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);

        if ($httpCode === 201) {
            $message = 'Device added successfully!';
        } else {
            $message = 'Failed to add device. Server response: ' . htmlspecialchars($response);
        }
        curl_close($ch);
    }
}
?>

<h2>Add a New Device</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>
<form method="post">
    <label for="label">Device label:</label>
    <input type="text" id="label" name="label" required><br><br>

    <label for="model">Model:</label>
    <select id="model" name="model" required>
        <option value="">-- Select --</option>
        <?php foreach ($models as $m): ?>
            <option value="<?= htmlspecialchars($m['model']) ?>">
                <?= htmlspecialchars($m['model']) ?>
            </option>
        <?php endforeach; ?>
    </select><br><br>

    <label for="zoneid">Zone:</label>
    <select id="zoneid" name="zoneid" required>
        <option value="">-- Select --</option>
        <?php foreach ($zones as $z): ?>
            <option value="<?= htmlspecialchars($z['id']) ?>">
                <?= htmlspecialchars($z['name']) ?>
            </option>
        <?php endforeach; ?>
    </select><br><br>

    <label for="proprietary">Proprietary:</label>
    <select id="proprietary" name="proprietary" required>
        <option value="">-- Select --</option>
        <?php foreach ($proprietaries as $prop): ?>
            <option value="<?= htmlspecialchars($prop['name']) ?>">
                <?= htmlspecialchars($prop['name']) ?>
            </option>
        <?php endforeach; ?>
    </select><br><br>

    <label for="profile">Scan profile (optional):</label>
    <select id="profile" name="profile">
        <option value="">-- None --</option>
        <?php foreach ($scanProfiles as $prof): $pk = ($prof['kind'] ?? '') !== '' ? $prof['kind'] : 'device'; ?>
            <option value="<?= htmlspecialchars($prof['name'] ?? '') ?>">
                <?= htmlspecialchars(($prof['name'] ?? '') . ' (' . $pk . (($prof['host'] ?? '') !== '' ? ', ' . $prof['host'] : '') . ')') ?>
            </option>
        <?php endforeach; ?>
    </select><br><br>

    <fieldset>
        <legend>IP Addresses (Optional)</legend>
        <div id="ipsContainer">
            <!-- IP addresses will be added here dynamically -->
        </div>
        <button type="button" onclick="addIPField()">+ Add IP Address</button>
    </fieldset>
    <br>

    <button type="submit">Add Device</button>
</form>

<script>
    let ipIndex = 0;

    function addIPField() {
        const container = document.getElementById('ipsContainer');
        const ipDiv = document.createElement('div');
        ipDiv.id = 'ip_' + ipIndex;
        ipDiv.style.marginBottom = '5px';

        ipDiv.innerHTML = `
            <input type="text" name="ips[]" placeholder="e.g., 192.168.1.100"
                   pattern="^(?:[0-9]{1,3}\\.){3}[0-9]{1,3}$"
                   title="IPv4 format: xxx.xxx.xxx.xxx"
                   style="width: 200px; margin-right: 5px;">
            <button type="button" onclick="removeIPField(${ipIndex})">Remove</button>
        `;

        container.appendChild(ipDiv);
        ipIndex++;
    }

    function removeIPField(index) {
        const element = document.getElementById('ip_' + index);
        if (element) {
            element.remove();
        }
    }
</script>
