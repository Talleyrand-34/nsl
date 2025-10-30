<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch models for selection
$models = json_decode(@file_get_contents(MODELS_ENDPOINT), true) ?: [];

// Default form values
$name = '';
$posx = '';
$posy = '';
$modelName = '';
$allowMultiple = false;

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
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
            'name' => $name,
            'posx' => $posx,
            'posy' => $posy,
            'modelName' => $modelName,
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

        if ($httpCode === 201) {
            $message = 'Model port added successfully!';
            // Do NOT reset the form values here
        } else {
            $message = 'Failed to add model port. Server response: ' . htmlspecialchars($response);
        }
        curl_close($ch);
    }
}
?>

<h2>Add a New Model Port</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>
<form method="post">
    <label for="name">Port name:</label>
    <input type="text" id="name" name="name" required value="<?= htmlspecialchars($name) ?>"><br><br>

    <label for="posx">Position X:</label>
    <input type="number" id="posx" name="posx" required value="<?= htmlspecialchars($posx) ?>"><br><br>

    <label for="posy">Position Y:</label>
    <input type="number" id="posy" name="posy" required value="<?= htmlspecialchars($posy) ?>"><br><br>

    <label for="modelName">Model:</label>
    <select id="modelName" name="modelName" required>
        <option value="">-- Select --</option>
        <?php foreach ($models as $model): ?>
            <option value="<?= htmlspecialchars($model['model']) ?>"
                <?= ($model['model'] === $modelName) ? 'selected' : '' ?>>
                <?= htmlspecialchars($model['model']) ?>
            </option>
        <?php endforeach; ?>
    </select><br><br>

    <label for="allow_multiple_connections">
        <input type="checkbox" id="allow_multiple_connections" name="allow_multiple_connections"
               <?= $allowMultiple ? 'checked' : '' ?>>
        Allow multiple connections to this port
    </label><br><br>

    <button type="submit">Add Model Port</button>
</form>
