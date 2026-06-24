<?php
require_once __DIR__ . '/config.php';

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

$pageTitle = 'NSL-Graph — Dashboard';
include __DIR__ . '/header.php';
?>
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
                <?php if (isset($_GET['allports'])): ?>
                    <input type="hidden" name="allports" value="<?= htmlspecialchars($_GET['allports']) ?>">
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
            <option value="false" <?= (isset($_GET['vlan']) && $_GET['vlan'] === 'false') ? 'selected' : '' ?>>No VLAN</option>
            <option value="true" <?= (!isset($_GET['vlan']) || $_GET['vlan'] === 'true') ? 'selected' : '' ?>>With VLANs</option>
        </select>

        <label for="colorPortsSelect" class="format-label" style="margin-left: 20px;">Color Ports with VLANs:</label>
        <select id="colorPortsSelect" name="colorports" onchange="this.form.submit()">
            <option value="false" <?= (isset($_GET['colorports']) && $_GET['colorports'] === 'false') ? 'selected' : '' ?>>No</option>
            <option value="true" <?= (!isset($_GET['colorports']) || $_GET['colorports'] === 'true') ? 'selected' : '' ?>>Yes</option>
        </select>

        <label for="allPortsSelect" class="format-label" style="margin-left: 20px;">Show All Ports:</label>
        <select id="allPortsSelect" name="allports" onchange="this.form.submit()">
            <option value="false" <?= (!isset($_GET['allports']) || $_GET['allports'] === 'false') ? 'selected' : '' ?>>No (only connected)</option>
            <option value="true" <?= (isset($_GET['allports']) && $_GET['allports'] === 'true') ? 'selected' : '' ?>>Yes (all ports)</option>
        </select>
    </div>

    <div class="diagram">
        <div class="resizable-img-container" style="height:500px;">
            <?php
                $format = isset($_GET['format']) ? htmlspecialchars($_GET['format']) : 'connections';
                $vlan = isset($_GET['vlan']) ? htmlspecialchars($_GET['vlan']) : 'true';
                $colorports = isset($_GET['colorports']) ? htmlspecialchars($_GET['colorports']) : 'true';
                $allports = isset($_GET['allports']) ? htmlspecialchars($_GET['allports']) : 'false';
                $diagramUrl = API_BASE_URL . "/diagram?format=" . $format . "&vlan=" . $vlan . "&colorports=" . $colorports . "&allports=" . $allports;
                // The server returns a non-2xx when the diagram can't be rendered
                // (e.g. no devices); the browser then shows the img's alt text.
            ?>
            <img id="diagramImg" src="<?= $diagramUrl ?>" alt="Diagram empty: No devices">
        </div>
    </div>
</form>
    </div>
<?php include __DIR__ . '/footer.php'; ?>
