<?php
require_once 'config.php';

// Set defaults
$actionType = $_GET['actionType'] ?? 'get';
$entity = $_GET['entity'] ?? 'brand';
$selectedAction = $actionType . $entity;

// For your action mapping:
$actionFiles = [
    'getbrand' => 'action/getbrand.php',
    'addbrand' => 'action/addbrand.php',
    'getdevclass' => 'action/getdevclass.php',
    'adddevclass' => 'action/adddevclass.php',
    'getproprietary' => 'action/getproprietary.php',
    'addproprietary' => 'action/addproprietary.php',
    'getzonetype' => 'action/getzonetype.php',
    'addzonetype' => 'action/addzonetype.php',
    'getzone' => 'action/getzone.php',
    'addzone' => 'action/addzone.php',
    'getmodeldevice' => 'action/getmodeldevice.php',
    'addmodeldevice' => 'action/addmodeldevice.php',
    'getdevice' => 'action/getdevice.php',
    'adddevice' => 'action/adddevice.php',
    'getmodelport' => 'action/getmodelport.php',
    'addmodelport' => 'action/addmodelport.php',
    'getconnections' => 'action/getconnections.php',
    'addconnections' => 'action/addconnections.php',
    'getconnectiontype' => 'action/getconnectiontype.php',
    'addconnectiontype' => 'action/addconnectiontype.php',
    'getdeviceport' => 'action/getdeviceport.php',
    'adddeviceport' => 'action/adddeviceport.php',
    // UPDATE actions
    'updatebrand' => 'action/updatebrand.php',
    'updatedevclass' => 'action/updatedevclass.php',
    'updateproprietary' => 'action/updateproprietary.php',
    'updatezonetype' => 'action/updatezonetype.php',
    'updatezone' => 'action/updatezone.php',
    'updatemodeldevice' => 'action/updatemodeldevice.php',
    'updatedevice' => 'action/updatedevice.php',
    'updatedeviceport' => 'action/updatedeviceport.php',
    'updatemodelport' => 'action/updatemodelport.php',
    'updateconnections' => 'action/updateconnections.php',
    'updateconnectiontype' => 'action/updateconnectiontype.php',
    // DELETE actions
    'deletebrand' => 'action/deletebrand.php',
    'deletedevclass' => 'action/deletedevclass.php',
    'deleteproprietary' => 'action/deleteproprietary.php',
    'deletezonetype' => 'action/deletezonetype.php',
    'deletezone' => 'action/deletezone.php',
    'deletemodeldevice' => 'action/deletemodeldevice.php',
    'deletedevice' => 'action/deletedevice.php',
    'deletemodelport' => 'action/deletemodelport.php',
    'deletedeviceport' => 'action/deletedeviceport.php',
    'deleteconnections' => 'action/deleteconnections.php',
    'deleteconnectiontype' => 'action/deleteconnectiontype.php',
    // VLAN actions
    'getvlan' => 'action/getvlan.php',
    'addvlan' => 'action/addvlan.php',
    'updatevlan' => 'action/updatevlan.php',
    'deletevlan' => 'action/deletevlan.php',
    // Add more as needed
];
?>
<?php
require_once 'config.php';

// Handle API base URL update
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['api_base_url'])) {
    $_SESSION['api_base_url'] = rtrim($_POST['api_base_url'], '/');  // remove trailing slash
    header("Location: " . $_SERVER['PHP_SELF'] . '?' . $_SERVER['QUERY_STRING']);
    exit;
}
?>
<!DOCTYPE html>
<html>
<head>
    <title>Brand Manager</title>
    <style>
        .grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 24px;
            align-items: start;
        }
        .actions {
            padding: 16px;
            border: 1px solid #ccc;
            border-radius: 6px;
            background: #f9f9f9;
            resize: horizontal;
            overflow: auto;
            min-width: 150px;
            max-width: 80vw;
        }
        .grid {
            grid-template-columns: auto 1fr;
        }
        .diagram {
            text-align: center;
        }
        .diagram img {
            max-width: 100%;
            height: auto;
            border: 1px solid #ccc;
            border-radius: 6px;
            display: block;
        }
        select, button {
            font-size: 1em;
            margin-top: 8px;
        }
        <!-- image resizing -->
        @property --img-width {
            syntax: "<length>";
            inherits: true;
            initial-value: 500px;
        }
        .diagram img {
            width: var(--img-width, 500px);
            height: auto;
        }
        input[type="range"] {
            width: 100%;
        }
        .resizable-img-container {
            resize: vertical;
            overflow: auto;
            min-height: 100px;   /* set as you wish */
            max-height: 190vh;    /* set as you wish */
            width: 100%;         /* or a fixed width if you prefer */
            border: 1px solid #ccc;
            display: flex;
            align-items: stretch;
            justify-content: center;
            background: #fff;
        }
        .resizable-img-container img {
            height: 100%;
            width: auto;
            display: block;
            object-fit: contain;
        }
    </style>
</head>
<body>
    <!-- Configuration Section -->
<div style="margin-top: 20px; padding: 10px; border: 1px solid #ccc; background: #f0f0f0;">
    <form method="post">
        <label for="apiUrl">API Base URL:</label>
        <input type="text" id="apiUrl" name="api_base_url" value="<?= htmlspecialchars(API_BASE_URL) ?>" style="width: 70%;">
        <button type="submit">Update</button>
    </form>
</div>
    <div class="grid">
        <div class="actions">
            <form method="get" action="">
                <!-- Hidden inputs to preserve diagram settings -->
                <?php if (isset($_GET['format'])): ?>
                    <input type="hidden" name="format" value="<?= htmlspecialchars($_GET['format']) ?>">
                <?php endif; ?>
                <?php if (isset($_GET['vlan'])): ?>
                    <input type="hidden" name="vlan" value="<?= htmlspecialchars($_GET['vlan']) ?>">
                <?php endif; ?>
                <?php if (isset($_GET['colorports'])): ?>
                    <input type="hidden" name="colorports" value="<?= htmlspecialchars($_GET['colorports']) ?>">
                <?php endif; ?>

                <label for="actionType">Choose action:</label>
                <select id="actionType" name="actionType" onchange="this.form.submit()">
                    <option value="get" <?= $actionType == 'get' ? 'selected' : '' ?>>Get</option>
                    <option value="add" <?= $actionType == 'add' ? 'selected' : '' ?>>Add</option>
                    <option value="update" <?= $actionType == 'update' ? 'selected' : '' ?>>Update</option>
                    <option value="delete" <?= $actionType == 'delete' ? 'selected' : '' ?>>Delete</option>
                </select>

                <label for="entity">Choose entity:</label>
                <select id="entity" name="entity" onchange="this.form.submit()">
                    <option value="brand" <?= $entity == 'brand' ? 'selected' : '' ?>>Brand</option>
                    <option value="devclass" <?= $entity == 'devclass' ? 'selected' : '' ?>>Device Class</option>
                    <option value="proprietary" <?= $entity == 'proprietary' ? 'selected' : '' ?>>Proprietary</option>
                    <option value="zonetype" <?= $entity == 'zonetype' ? 'selected' : '' ?>>Zone Type</option>
                    <option value="zone" <?= $entity == 'zone' ? 'selected' : '' ?>>Zone</option>
                    <option value="modeldevice" <?= $entity == 'modeldevice' ? 'selected' : '' ?>>Model</option>
                    <option value="device" <?= $entity == 'device' ? 'selected' : '' ?>>Device</option>
                    <option value="modelport" <?= $entity == 'modelport' ? 'selected' : '' ?>>ModelPort</option>
                    <option value="deviceport" <?= $entity == 'deviceport' ? 'selected' : '' ?>>DevicePort</option>
                    <option value="connections" <?= $entity == 'connections' ? 'selected' : '' ?>>connections</option>
                    <option value="connectiontype" <?= $entity == 'connectiontype' ? 'selected' : '' ?>>connectiontypes</option>
                    <option value="vlan" <?= $entity == 'vlan' ? 'selected' : '' ?>>VLAN</option>
                    <!-- Add more entities as needed -->
                </select>
                <noscript><button type="submit">Go</button></noscript>
            </form>
            <div style="margin-top:20px;">
                <?php
                if (isset($actionFiles[$selectedAction])) {
                    include $actionFiles[$selectedAction];
                } else {
                    echo '<em>Select an action from the menu.</em>';
                }
                ?>
            </div>
        </div>

            <form method="GET" action="">
    <!-- Hidden inputs to preserve current state -->
    <input type="hidden" name="actionType" value="<?= htmlspecialchars($actionType) ?>">
    <input type="hidden" name="entity" value="<?= htmlspecialchars($entity) ?>">

    <div class="diagram-controls">
        <label for="formatSelect" class="format-label">Diagram Format:</label>
        <select id="formatSelect" name="format" onchange="this.form.submit()">
            <option value="ports" <?= (isset($_GET['format']) && $_GET['format'] === 'ports') ? 'selected' : '' ?>>Ports</option>
            <option value="connections" <?= (!isset($_GET['format']) || $_GET['format'] === 'connections') ? 'selected' : '' ?>>Connections</option>
        </select>

        <label for="vlanSelect" class="format-label" style="margin-left: 20px;">VLAN Display:</label>
        <select id="vlanSelect" name="vlan" onchange="this.form.submit()">
            <option value="false" <?= (!isset($_GET['vlan']) || $_GET['vlan'] === 'false') ? 'selected' : '' ?>>No VLAN</option>
            <option value="true" <?= (isset($_GET['vlan']) && $_GET['vlan'] === 'true') ? 'selected' : '' ?>>With VLANs</option>
        </select>

        <label for="colorPortsSelect" class="format-label" style="margin-left: 20px;">Color Ports with VLANs:</label>
        <select id="colorPortsSelect" name="colorports" onchange="this.form.submit()">
            <option value="false" <?= (!isset($_GET['colorports']) || $_GET['colorports'] === 'false') ? 'selected' : '' ?>>No</option>
            <option value="true" <?= (isset($_GET['colorports']) && $_GET['colorports'] === 'true') ? 'selected' : '' ?>>Yes</option>
        </select>
    </div>

    <!-- Plugin Selector -->
    <div style="margin: 20px 0; padding: 15px; background-color: #f5f5f5; border-radius: 4px;">
        <strong>Connection Sorting Plugin:</strong>
        <select id="pluginSelect" style="margin-left: 10px; padding: 5px;" onchange="changePlugin()">
            <option value="">Loading plugins...</option>
        </select>
        <span id="pluginStatus" style="margin-left: 10px; color: #666;"></span>
        <!-- <button type="button" onclick="refreshDiagram()" style="margin-left: 20px; padding: 5px 15px; background-color: #4CAF50; color: white; border: none; border-radius: 4px; cursor: pointer;"> -->
        <!--     🔄 Refresh Diagram -->
        <!-- </button> -->
        <p id="pluginDescription" style="margin-top: 10px; font-size: 0.9em; color: #555;"></p>
    </div>

    <div class="diagram">
        <div class="resizable-img-container" style="height:500px;">
            <?php
                $format = isset($_GET['format']) ? htmlspecialchars($_GET['format']) : 'connections';
                $vlan = isset($_GET['vlan']) ? htmlspecialchars($_GET['vlan']) : 'false';
                $colorports = isset($_GET['colorports']) ? htmlspecialchars($_GET['colorports']) : 'false';
                $diagramUrl = API_BASE_URL . "/diagram?format=" . $format . "&vlan=" . $vlan . "&colorports=" . $colorports;
            ?>
            <img id="diagramImg" src="<?= $diagramUrl ?>" alt="Diagram">
        </div>
    </div>
</form>

<script>
// Function to refresh diagram (called by button)
function refreshDiagram() {
    sessionStorage.setItem('refreshDiagram', 'true');
    location.reload();
}

// Plugin management functions
let currentPlugins = [];

// Load available plugins
function loadPlugins() {
    const pluginsUrl = '<?= PLUGINS_ENDPOINT ?>';
    console.log('Loading plugins from:', pluginsUrl);

    fetch(pluginsUrl)
        .then(response => {
            console.log('Response status:', response.status);
            if (!response.ok) {
                throw new Error('HTTP ' + response.status + ': ' + response.statusText);
            }
            return response.json();
        })
        .then(plugins => {
            console.log('Loaded plugins:', plugins);
            currentPlugins = plugins;
            const select = document.getElementById('pluginSelect');
            select.innerHTML = '';

            plugins.forEach(plugin => {
                const option = document.createElement('option');
                option.value = plugin.id;
                option.textContent = plugin.name;
                if (plugin.is_active) {
                    option.selected = true;
                    updatePluginDescription(plugin);
                }
                select.appendChild(option);
            });

            document.getElementById('pluginStatus').textContent = '';
        })
        .catch(error => {
            console.error('Error loading plugins:', error);
            const statusElem = document.getElementById('pluginStatus');
            statusElem.textContent = '❌ Failed to load plugins: ' + error.message;
            statusElem.style.color = 'red';

            const descElem = document.getElementById('pluginDescription');
            descElem.textContent = 'Endpoint: ' + pluginsUrl + ' - Check browser console for details';
            descElem.style.color = 'red';
        });
}

// Update plugin description display
function updatePluginDescription(plugin) {
    const descElem = document.getElementById('pluginDescription');
    if (plugin) {
        descElem.textContent = '📝 ' + plugin.description;
    } else {
        descElem.textContent = '';
    }
}

// Change active plugin
function changePlugin() {
    const select = document.getElementById('pluginSelect');
    const pluginId = select.value;
    const statusElem = document.getElementById('pluginStatus');

    // Find selected plugin info
    const plugin = currentPlugins.find(p => p.id === pluginId);
    if (plugin) {
        updatePluginDescription(plugin);
    }

    statusElem.textContent = '⏳ Changing plugin...';
    statusElem.style.color = '#666';

    fetch('<?= PLUGINS_ACTIVE_ENDPOINT ?>', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ plugin_id: pluginId })
    })
    .then(response => {
        if (!response.ok) {
            return response.text().then(text => {
                throw new Error(text || 'Failed to change plugin');
            });
        }
        return response.json();
    })
    .then(data => {
        statusElem.textContent = '✓ Plugin changed successfully';
        statusElem.style.color = 'green';

        // Clear status after 3 seconds
        setTimeout(() => {
            statusElem.textContent = '';
        }, 3000);

        // Note: No need to reload - next connection fetch will use new plugin
        // User can click refresh button if they want to see changes immediately
    })
    .catch(error => {
        console.error('Error changing plugin:', error);
        statusElem.textContent = '❌ ' + error.message;
        statusElem.style.color = 'red';

        // Reload plugins to restore correct selection
        loadPlugins();
    });
}

// Load plugins on page load
document.addEventListener('DOMContentLoaded', loadPlugins);
</script>
</body>
</html>
