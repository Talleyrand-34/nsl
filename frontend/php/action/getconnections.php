
<?php
require_once __DIR__ . '/../config.php';

// Fetch connections
$connectionsJson = @file_get_contents(CONNECTIONS_ENDPOINT);
$connections = json_decode($connectionsJson, true);

if (is_array($connections)) {
    echo '<ul>';
    foreach ($connections as $conn) {
        echo '<li>';
        echo '<strong>Connection ID: ' . htmlspecialchars($conn['id']) . '</strong><br>';
        echo 'From Device: <b>' . htmlspecialchars($conn['fromdevice']) . '</b> (Port: ' . htmlspecialchars($conn['frommodel']) . ')';
        if (!empty($conn['fromipsegment'])) {
            echo ' | IP: ' . htmlspecialchars($conn['fromipsegment']);
        }
        echo ' | Zone: ' . htmlspecialchars($conn['fromzonename']) . ' (ID: ' . htmlspecialchars($conn['fromzoneid']) . ')<br>';
        echo 'To Device: <b>' . htmlspecialchars($conn['todevice']) . '</b> (Port: ' . htmlspecialchars($conn['tomodel']) . ')';
        if (!empty($conn['toipsegment'])) {
            echo ' | IP: ' . htmlspecialchars($conn['toipsegment']);
        }
        echo ' | Zone: ' . htmlspecialchars($conn['tozonename']) . ' (ID: ' . htmlspecialchars($conn['tozoneid']) . ')';
        echo '</li><hr>';
    }
    echo '</ul>';
} else {
    echo '<p><em>Could not fetch connections.</em></p>';
}
?>
