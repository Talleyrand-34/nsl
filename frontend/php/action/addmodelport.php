<?php
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

require_once __DIR__ . '/../config.php';
$message = '';
$bulkMode = isset($_GET['bulk']) && $_GET['bulk'] === 'true';

// Fetch models for selection
$models = json_decode(@file_get_contents(MODELS_ENDPOINT), true) ?: [];

// Default form values
$name = '';
$posx = '';
$posy = '';
$modelName = '';
$allowMultiple = false;
$bulkData = '';
$bulkFormat = 'simple'; // simple, keyvalue, generate

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    if (isset($_POST['bulk_mode']) && $_POST['bulk_mode'] === 'true') {
        // Bulk mode processing
        $bulkData = trim($_POST['bulk_data'] ?? '');
        $modelName = $_POST['bulk_modelName'] ?? '';
        $allowMultiple = isset($_POST['bulk_allow_multiple_connections']);
        $bulkFormat = $_POST['bulk_format'] ?? 'simple';

        if (empty($modelName)) {
            $message = 'Please select a model.';
        } elseif ($bulkFormat === 'generate') {
            // Auto-generate mode
            $prefix = trim($_POST['port_prefix'] ?? '');
            $count = intval($_POST['port_count'] ?? 0);
            $startIdx = intval($_POST['start_index'] ?? 0);
            $startX = intval($_POST['start_x'] ?? 0);
            $startY = intval($_POST['start_y'] ?? 0);
            $incrX = intval($_POST['incr_x'] ?? 1);
            $incrY = intval($_POST['incr_y'] ?? 0);

            if (empty($prefix)) {
                $message = 'Please enter a port name prefix.';
            } elseif ($count < 1 || $count > 100) {
                $message = 'Port count must be between 1 and 100.';
            } else {
                $ports = [];
                for ($i = 0; $i < $count; $i++) {
                    $ports[] = [
                        'name' => $prefix . ($startIdx + $i),
                        'posx' => (string)($startX + ($i * $incrX)),
                        'posy' => (string)($startY + ($i * $incrY)),
                        'modelName' => $modelName,
                        'allow_multiple_connections' => $allowMultiple
                    ];
                }

                // Send bulk request
                $result = sendBulkRequest($ports);
                $message = $result['message'];
            }
        } elseif (empty($bulkData)) {
            $message = 'Please enter port data.';
        } else {
            // Parse bulk data based on format
            $ports = [];
            $parseErrors = [];

            $lines = array_filter(array_map('trim', explode("\n", $bulkData)));

            foreach ($lines as $lineNum => $line) {
                $lineIndex = $lineNum + 1;

                if ($bulkFormat === 'simple') {
                    // Simple format: name,posx,posy or name,posx,posy,model
                    $parts = array_map('trim', explode(',', $line));

                    if (count($parts) < 3) {
                        $parseErrors[] = "Line {$lineIndex}: Expected at least 3 values (name,posx,posy)";
                        continue;
                    }

                    $portName = $parts[0];
                    $portX = $parts[1];
                    $portY = $parts[2];
                    $lineModel = count($parts) > 3 ? $parts[3] : $modelName;

                    if (empty($portName)) {
                        $parseErrors[] = "Line {$lineIndex}: Port name cannot be empty";
                        continue;
                    }

                    if (!is_numeric($portX) || !is_numeric($portY)) {
                        $parseErrors[] = "Line {$lineIndex}: Invalid position values (must be numeric)";
                        continue;
                    }

                    $ports[] = [
                        'name' => $portName,
                        'posx' => (string)((int)$portX),
                        'posy' => (string)((int)$portY),
                        'modelName' => $lineModel,
                        'allow_multiple_connections' => $allowMultiple
                    ];

                } elseif ($bulkFormat === 'keyvalue') {
                    // Key-value format: name:value,posx:value,posy:value
                    $parts = array_filter(array_map('trim', explode(',', $line)));
                    $portData = [];

                    foreach ($parts as $part) {
                        $kv = array_map('trim', explode(':', $part, 2));
                        if (count($kv) === 2) {
                            $portData[strtolower($kv[0])] = $kv[1];
                        }
                    }

                    if (!isset($portData['name'])) {
                        $parseErrors[] = "Line {$lineIndex}: Missing 'name' field";
                        continue;
                    }

                    $portName = $portData['name'];
                    $portX = $portData['posx'] ?? $portData['positionx'] ?? '0';
                    $portY = $portData['posy'] ?? $portData['positiony'] ?? '0';
                    $lineModel = $portData['model'] ?? $portData['modelname'] ?? $modelName;

                    if (!is_numeric($portX) || !is_numeric($portY)) {
                        $parseErrors[] = "Line {$lineIndex}: Invalid position values (must be numeric)";
                        continue;
                    }

                    $ports[] = [
                        'name' => $portName,
                        'posx' => (string)((int)$portX),
                        'posy' => (string)((int)$portY),
                        'modelName' => $lineModel,
                        'allow_multiple_connections' => $allowMultiple
                    ];
                }
            }

            if (!empty($parseErrors)) {
                $message = 'Parsing errors:<br>' . implode('<br>', $parseErrors);
            } elseif (empty($ports)) {
                $message = 'No valid ports to create.';
            } else {
                // Send bulk request
                $result = sendBulkRequest($ports);
                $message = $result['message'];
            }
        }
    } else {
        // Single mode processing
        $name = trim($_POST['name'] ?? '');
        $posx = trim($_POST['posx'] ?? '');
        $posy = trim($_POST['posy'] ?? '');
        $modelName = $_POST['modelName'] ?? '';
        $allowMultiple = isset($_POST['allow_multiple_connections']);

        if ($name === '') {
            $message = 'Please enter a port name.';
        } elseif ($posx === '' || !is_numeric($posx)) {
            $message = 'Please enter a valid X position.';
        } elseif ($posy === '' || !is_numeric($posy)) {
            $message = 'Please enter a valid Y position.';
        } elseif ($modelName === '') {
            $message = 'Please select a model.';
        } else {
            $data = json_encode([
                'port_name' => $name,
                'position_x' => $posx,
                'position_y' => $posy,
                'model_name' => $modelName,
                'allow_multiple_connections' => $allowMultiple
            ]);

            $ch = curl_init(MODELPORTS_ENDPOINT);
            curl_setopt($ch, CURLOPT_CUSTOMREQUEST, 'POST');
            curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
            curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
            curl_setopt($ch, CURLOPT_HTTPHEADER, [
                'Content-Type: application/json',
                'Content-Length: ' . strlen($data)
            ]);

            $response = curl_exec($ch);
            $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
            curl_close($ch);

            if ($httpCode === 201) {
                $message = 'Model port added successfully!';
                // Clear form on success
                $name = $posx = $posy = $modelName = '';
                $allowMultiple = false;
            } else {
                $message = 'Failed to add model port. Server response: ' . htmlspecialchars($response);
            }
        }
    }
}

function sendBulkRequest($ports) {
    $data = json_encode(['ports' => $ports]);

    $ch = curl_init(MODELPORTS_BULK_ENDPOINT);
    curl_setopt($ch, CURLOPT_CUSTOMREQUEST, 'POST');
    curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_HTTPHEADER, [
        'Content-Type: application/json',
        'Content-Length: ' . strlen($data)
    ]);

    $response = curl_exec($ch);
    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);

    if ($httpCode === 201 || $httpCode === 207) {
        $result = json_decode($response, true);
        $successCount = $result['successCount'] ?? 0;
        $failureCount = $result['failureCount'] ?? 0;
        $errors = $result['errors'] ?? [];

        if ($successCount > 0 && $failureCount == 0) {
            return ['success' => true, 'message' => "Successfully created {$successCount} port(s)!"];
        } elseif ($successCount > 0) {
            return ['success' => true, 'message' => "Partially successful: Created {$successCount} port(s), failed {$failureCount}.<br>Errors:<br>" . implode('<br>', $errors)];
        } else {
            return ['success' => false, 'message' => "Failed to create any ports.<br>Errors:<br>" . implode('<br>', $errors)];
        }
    } else {
        return ['success' => false, 'message' => 'Failed to create ports. Server response: ' . htmlspecialchars($response)];
    }
}
?>

<!DOCTYPE html>
<html>
<head>
    <style>
        .form-section { margin-bottom: 20px; }
        .help-text { font-size: 0.9em; color: #666; margin-top: 5px; }
        .format-example { background: #f5f5f5; padding: 10px; margin: 10px 0; border-left: 3px solid #007bff; }
        .format-example code { display: block; margin: 5px 0; }
        .message { padding: 10px; margin: 10px 0; border-radius: 4px; }
        .message.success { background: #d4edda; border: 1px solid #c3e6cb; }
        .message.error { background: #f8d7da; border: 1px solid #f5c6cb; }
        .tab-buttons { margin: 10px 0; }
        .tab-buttons button { padding: 8px 15px; margin-right: 5px; cursor: pointer; }
        .tab-buttons button.active { background: #007bff; color: white; border: none; }
        .tab-content { display: none; }
        .tab-content.active { display: block; }
        input[type="text"], input[type="number"], select, textarea {
            padding: 5px;
            margin: 5px 0;
        }
        textarea { font-family: monospace; }
    </style>
</head>
<body>

<h2>
    <?= $bulkMode ? 'Bulk Add Model Ports' : 'Add a New Model Port' ?>
    (<a href="?actionType=add&entity=modelport&bulk=<?= $bulkMode ? 'false' : 'true' ?>"><?= $bulkMode ? 'Switch to Single Mode' : 'Switch to Bulk Mode' ?></a>)
</h2>

<?php if ($message): ?>
    <div class="message <?= strpos($message, 'Successfully') !== false || strpos($message, 'success') !== false ? 'success' : 'error' ?>">
        <?= $message ?>
    </div>
<?php endif; ?>

<?php if ($bulkMode): ?>
    <!-- Bulk Mode Form -->
    <form method="post" id="bulkForm">
        <input type="hidden" name="bulk_mode" value="true">
        <input type="hidden" name="bulk_format" id="bulk_format" value="<?= htmlspecialchars($bulkFormat) ?>">

        <div class="form-section">
            <label for="bulk_modelName"><strong>Model:</strong></label>
            <select id="bulk_modelName" name="bulk_modelName" required>
                <option value="">-- Select Model --</option>
                <?php foreach ($models as $model): ?>
                    <option value="<?= htmlspecialchars($model['model']) ?>"
                        <?= ($model['model'] === $modelName) ? 'selected' : '' ?>>
                        <?= htmlspecialchars($model['model']) ?>
                    </option>
                <?php endforeach; ?>
            </select>
        </div>

        <div class="form-section">
            <label>
                <input type="checkbox" name="bulk_allow_multiple_connections" <?= $allowMultiple ? 'checked' : '' ?>>
                Allow multiple connections to all ports
            </label>
        </div>

        <hr>

        <div class="tab-buttons">
            <button type="button" onclick="switchTab('simple')" id="tab-simple" class="active">Simple Format</button>
            <button type="button" onclick="switchTab('keyvalue')" id="tab-keyvalue">Key-Value Format</button>
            <button type="button" onclick="switchTab('generate')" id="tab-generate">Auto-Generate</button>
        </div>

        <!-- Simple Format Tab -->
        <div id="content-simple" class="tab-content active">
            <div class="format-example">
                <strong>Format:</strong> <code>name,posx,posy</code> (one per line)<br>
                <strong>Example:</strong>
                <code>eth0,0,0</code>
                <code>eth1,1,0</code>
                <code>eth2,2,0</code>
            </div>

            <label for="bulk_data_simple"><strong>Port Data:</strong></label>
            <textarea id="bulk_data_simple" name="bulk_data" rows="12" cols="60" placeholder="eth0,0,0&#10;eth1,1,0&#10;eth2,2,0"><?= $bulkFormat === 'simple' ? htmlspecialchars($bulkData) : '' ?></textarea>
            <div class="help-text">Enter one port per line in format: name,x-position,y-position</div>
        </div>

        <!-- Key-Value Format Tab -->
        <div id="content-keyvalue" class="tab-content">
            <div class="format-example">
                <strong>Format:</strong> <code>name:value,posx:value,posy:value</code><br>
                <strong>Example:</strong>
                <code>name:GigE1/0/1,posx:0,posy:0</code>
                <code>name:GigE1/0/2,posx:1,posy:0</code>
            </div>

            <label for="bulk_data_keyvalue"><strong>Port Data:</strong></label>
            <textarea id="bulk_data_keyvalue" name="bulk_data_kv" rows="12" cols="60" placeholder="name:GigE1/0/1,posx:0,posy:0&#10;name:GigE1/0/2,posx:1,posy:0"><?= $bulkFormat === 'keyvalue' ? htmlspecialchars($bulkData) : '' ?></textarea>
            <div class="help-text">Enter one port per line. Supports: name, posx/positionx, posy/positiony, model/modelname</div>
        </div>

        <!-- Auto-Generate Tab -->
        <div id="content-generate" class="tab-content">
            <div class="format-example">
                <strong>Automatically generate multiple ports with sequential naming</strong>
            </div>

            <div class="form-section">
                <label for="port_prefix"><strong>Port Name Prefix:</strong></label>
                <input type="text" id="port_prefix" name="port_prefix" placeholder="eth" value="eth">
                <div class="help-text">Prefix for port names (e.g., "eth" generates eth0, eth1, eth2...)</div>
            </div>

            <div class="form-section">
                <label for="port_count"><strong>Number of Ports:</strong></label>
                <input type="number" id="port_count" name="port_count" value="28" min="1" max="100">
            </div>

            <div class="form-section">
                <label for="start_index"><strong>Starting Index:</strong></label>
                <input type="number" id="start_index" name="start_index" value="0" min="0">
                <div class="help-text">First port will be: prefix + this number</div>
            </div>

            <div class="form-section">
                <label for="start_x"><strong>Starting X Position:</strong></label>
                <input type="number" id="start_x" name="start_x" value="0">
            </div>

            <div class="form-section">
                <label for="start_y"><strong>Starting Y Position:</strong></label>
                <input type="number" id="start_y" name="start_y" value="0">
            </div>

            <div class="form-section">
                <label for="incr_x"><strong>X Increment:</strong></label>
                <input type="number" id="incr_x" name="incr_x" value="1">
                <div class="help-text">How much to increase X for each port</div>
            </div>

            <div class="form-section">
                <label for="incr_y"><strong>Y Increment:</strong></label>
                <input type="number" id="incr_y" name="incr_y" value="0">
                <div class="help-text">How much to increase Y for each port</div>
            </div>
        </div>

        <button type="submit">Create Ports</button>
    </form>

    <script>
        function switchTab(tab) {
            // Update hidden field
            document.getElementById('bulk_format').value = tab;

            // Update tab buttons
            document.querySelectorAll('.tab-buttons button').forEach(btn => {
                btn.classList.remove('active');
            });
            document.getElementById('tab-' + tab).classList.add('active');

            // Update tab content
            document.querySelectorAll('.tab-content').forEach(content => {
                content.classList.remove('active');
            });
            document.getElementById('content-' + tab).classList.add('active');

            // Copy textarea content if switching between simple and keyvalue
            if (tab === 'simple') {
                var kvData = document.getElementById('bulk_data_keyvalue').value;
                if (kvData) {
                    document.getElementById('bulk_data_simple').value = kvData;
                }
            } else if (tab === 'keyvalue') {
                var simpleData = document.getElementById('bulk_data_simple').value;
                if (simpleData) {
                    document.getElementById('bulk_data_keyvalue').value = simpleData;
                }
            }
        }

        // Handle form submission to use the correct textarea
        document.getElementById('bulkForm').addEventListener('submit', function(e) {
            var format = document.getElementById('bulk_format').value;
            if (format === 'keyvalue') {
                // Copy keyvalue textarea to bulk_data
                var kvData = document.getElementById('bulk_data_keyvalue').value;
                document.getElementById('bulk_data_simple').value = kvData;
                document.getElementById('bulk_data_simple').name = 'bulk_data';
            }
        });
    </script>

<?php else: ?>
    <!-- Single Mode Form -->
    <form method="post">
        <div class="form-section">
            <label for="name"><strong>Port Name:</strong></label>
            <input type="text" id="name" name="name" required value="<?= htmlspecialchars($name) ?>" placeholder="eth0">
        </div>

        <div class="form-section">
            <label for="posx"><strong>Position X:</strong></label>
            <input type="number" id="posx" name="posx" required value="<?= htmlspecialchars($posx) ?>" placeholder="0">
        </div>

        <div class="form-section">
            <label for="posy"><strong>Position Y:</strong></label>
            <input type="number" id="posy" name="posy" required value="<?= htmlspecialchars($posy) ?>" placeholder="0">
        </div>

        <div class="form-section">
            <label for="modelName"><strong>Model:</strong></label>
            <select id="modelName" name="modelName" required>
                <option value="">-- Select Model --</option>
                <?php foreach ($models as $model): ?>
                    <option value="<?= htmlspecialchars($model['model']) ?>"
                        <?= ($model['model'] === $modelName) ? 'selected' : '' ?>>
                        <?= htmlspecialchars($model['model']) ?>
                    </option>
                <?php endforeach; ?>
            </select>
        </div>

        <div class="form-section">
            <label>
                <input type="checkbox" id="allow_multiple_connections" name="allow_multiple_connections"
                       <?= $allowMultiple ? 'checked' : '' ?>>
                Allow multiple connections to this port
            </label>
        </div>

        <button type="submit">Add Model Port</button>
    </form>
<?php endif; ?>

</body>
</html>
