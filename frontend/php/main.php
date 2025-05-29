<?php
require_once 'config.php';

// // Map enum values to PHP files
// $actionFiles = [
//     'getbrand'  => 'action/getbrand.php',
//     'addbrand'  => 'action/addbrand.php',
//     'getdevclass'  => 'action/getdevclass.php',
//     'getproprietary'  => 'action/getproprietary.php',
//     'getzonetype'  => 'action/getzonetype.php',
//     'getzone'  => 'action/getzone.php',
//     // Add more actions as needed
// ];
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
    // Add more as needed
];
// $selectedAction = $_GET['action'] ?? 'getbrand';
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
    <div class="grid">
        <div class="actions">
            <form method="get" action="">
                <label for="actionType">Choose action:</label>
                <select id="actionType" name="actionType" onchange="this.form.submit()">
                    <option value="get" <?= $actionType == 'get' ? 'selected' : '' ?>>Get</option>
                    <option value="add" <?= $actionType == 'add' ? 'selected' : '' ?>>Add</option>
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
                    <option value="connections" <?= $entity == 'connections' ? 'selected' : '' ?>>connections</option>
                    <option value="connectiontype" <?= $entity == 'connectiontype' ? 'selected' : '' ?>>connectiontypes</option>
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
        <!-- <div class="diagram" style="--img-width:500px;"> -->
        <!--         <img id="diagramImg" src="<?= API_BASE_URL ?>/diagram" alt="Diagram" style="width:500px;"> -->
        <!-- </div> -->
        <div class="diagram">
            <div class="resizable-img-container" style="height:500px;">
                <img id="diagramImg" src="<?= API_BASE_URL ?>/diagram" alt="Diagram">
            </div>
        </div>

</body>
</html>
