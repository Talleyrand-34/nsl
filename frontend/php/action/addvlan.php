<?php
require_once __DIR__ . '/../config.php';
$message = '';

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $vlanID = trim($_POST['vlan_id'] ?? '');
    $vlanName = trim($_POST['vlan_name'] ?? '');

    // Collect IP segment IDs from dynamic inputs
    $ipSegmentIDs = [];
    if (isset($_POST['ip_segment_ids']) && is_array($_POST['ip_segment_ids'])) {
        foreach ($_POST['ip_segment_ids'] as $segmentId) {
            $segmentId = trim($segmentId);
            if (!empty($segmentId)) {
                $ipSegmentIDs[] = $segmentId;
            }
        }
    }

    if ($vlanID !== '') {
        $data = json_encode([
            'vlanID' => $vlanID,
            'vlanName' => $vlanName,
            'ipSegmentIDs' => $ipSegmentIDs
        ]);

        $ch = curl_init(VLANS_ENDPOINT);
        curl_setopt($ch, CURLOPT_POST, true);
        curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'Content-Type: application/json',
            'Content-Length: ' . strlen($data)
        ]);

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);

        if ($httpCode === 201 || $httpCode === 200) {
            $message = "VLAN created successfully!";
        } else {
            $message = "Failed to create VLAN. Server response: " . htmlspecialchars($response);
        }
        curl_close($ch);
    } else {
        $message = "VLAN ID is required.";
    }
}
?>

<h2>Add a New VLAN</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<form method="post">
    <label for="vlan_id">VLAN ID (e.g., 100):</label>
    <input type="text" id="vlan_id" name="vlan_id" required><br><br>

    <label for="vlan_name">VLAN Name (optional, e.g., "Management VLAN"):</label>
    <input type="text" id="vlan_name" name="vlan_name"><br><br>

    <fieldset>
        <legend>IP Segment IDs (Optional)</legend>
        <div id="ipSegmentIDsContainer">
            <!-- IP segment IDs will be added here dynamically -->
        </div>
        <button type="button" onclick="addIPSegmentField()">+ Add IP Segment ID</button>
        <br><small><em>Add IP segment IDs that are associated with this VLAN</em></small>
    </fieldset>
    <br>

    <button type="submit">Add VLAN</button>
</form>

<script>
    let ipSegmentIndex = 0;

    function addIPSegmentField() {
        const container = document.getElementById('ipSegmentIDsContainer');
        const segmentDiv = document.createElement('div');
        segmentDiv.id = 'ipSegment_' + ipSegmentIndex;
        segmentDiv.style.marginBottom = '5px';

        segmentDiv.innerHTML = `
            <input type="text" name="ip_segment_ids[]" placeholder="e.g., segment-001"
                   style="width: 250px; margin-right: 5px;">
            <button type="button" onclick="removeIPSegmentField(${ipSegmentIndex})">Remove</button>
        `;

        container.appendChild(segmentDiv);
        ipSegmentIndex++;
    }

    function removeIPSegmentField(index) {
        const element = document.getElementById('ipSegment_' + index);
        if (element) {
            element.remove();
        }
    }
</script>
