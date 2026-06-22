<?php
require_once __DIR__ . '/config.php';

// Shared "API Base URL" banner behaviour (same as main.php).
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['api_base_url'])) {
    $_SESSION['api_base_url'] = rtrim($_POST['api_base_url'], '/');
    header('Location: ' . $_SERVER['PHP_SELF']);
    exit;
}
?>
<!DOCTYPE html>
<html>
<head>
    <title>Import devices — NSL-Graph</title>
    <style>
        body { font-family: sans-serif; margin: 16px; }
        table { border-collapse: collapse; }
        h2, h3 { margin-top: 20px; }
        label { display: inline-block; margin: 4px 0; }
        /* Same look as the CRUD/diagram boxes in main.php */
        .grid2 {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 24px;
            align-items: start;
            margin-top: 8px;
        }
        .box {
            padding: 16px;
            border: 1px solid #ccc;
            border-radius: 6px;
            background: #f9f9f9;
            resize: horizontal;
            overflow: auto;
            min-width: 150px;
            max-width: 80vw;
        }
        .box h3 { margin-top: 0; }
    </style>
</head>
<body>
    <div style="padding: 8px 10px; border: 1px solid #ccc; background: #eef;">
        <a href="main.php">&larr; Home</a>
    </div>
    <div style="margin-top: 12px; padding: 10px; border: 1px solid #ccc; background: #f0f0f0;">
        <form method="post">
            <label for="apiUrl">API Base URL:</label>
            <input type="text" id="apiUrl" name="api_base_url" value="<?= htmlspecialchars(API_BASE_URL) ?>" style="width: 70%;">
            <button type="submit">Update</button>
        </form>
    </div>
    <div style="margin-top: 16px;">
<?php include __DIR__ . '/action/importscan.php'; ?>
    </div>
</body>
</html>
