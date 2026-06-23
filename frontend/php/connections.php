<?php
require_once __DIR__ . '/config.php';

// Shared "API Base URL" banner behaviour (same as main.php / import.php).
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['api_base_url'])) {
    $_SESSION['api_base_url'] = rtrim($_POST['api_base_url'], '/');
    header('Location: ' . $_SERVER['PHP_SELF']);
    exit;
}
?>
<!DOCTYPE html>
<html>
<head>
    <title>Scan connections — NSL-Graph</title>
    <style>
        body { font-family: sans-serif; margin: 16px; }
        table { border-collapse: collapse; }
        h2, h3 { margin-top: 20px; }
        label { display: block; margin: 6px 0; }
        .box {
            padding: 16px;
            border: 1px solid #ccc;
            border-radius: 6px;
            background: #f9f9f9;
        }
        .box h3 { margin-top: 0; }
    </style>
</head>
<body>
    <div style="padding: 8px 10px; border: 1px solid #ccc; background: #eef;">
        <a href="main.php">&larr; Home</a> &nbsp;|&nbsp;
        <a href="import.php">Import devices</a>
    </div>
    <div style="margin-top: 12px; padding: 10px; border: 1px solid #ccc; background: #f0f0f0;">
        <form method="post">
            <label for="apiUrl" style="display:inline;">API Base URL:</label>
            <input type="text" id="apiUrl" name="api_base_url" value="<?= htmlspecialchars(API_BASE_URL) ?>" style="width: 70%;">
            <button type="submit">Update</button>
        </form>
    </div>
    <div style="margin-top: 16px;">
<?php include __DIR__ . '/action/connections.php'; ?>
    </div>
</body>
</html>
