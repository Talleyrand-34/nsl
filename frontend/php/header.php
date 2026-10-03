<?php
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

// Shared page header: the API-base-URL handler, document <head>, and the app
// bar (section navigation + API-URL control) rendered identically on every page.
// A page sets an optional $pageTitle, then `include __DIR__.'/header.php';`.
require_once __DIR__ . '/config.php';

// Handle the API base URL update before any output, preserving the query string
// so a page's state (e.g. main.php's actionType/entity/diagram options) survives.
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['api_base_url'])) {
    $_SESSION['api_base_url'] = rtrim($_POST['api_base_url'], '/');
    $qs = $_SERVER['QUERY_STRING'] ?? '';
    header('Location: ' . $_SERVER['PHP_SELF'] . ($qs !== '' ? '?' . $qs : ''));
    exit;
}

// Top-level sections, in display order. The active one is auto-detected from the
// current script name, so pages need no manual flag.
$navItems = [
    'main.php'        => 'Dashboard',
    'push.php'        => 'Push config',
    'import.php'      => 'Import devices',
];
$currentPage = basename($_SERVER['PHP_SELF']);
?>
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title><?= htmlspecialchars($pageTitle ?? 'NSL-Graph') ?></title>
    <!-- Apply the saved theme before the stylesheet/body paint to avoid a flash. -->
    <script>
    (function () { try { if (localStorage.getItem('nsl-theme') === 'dark') document.documentElement.setAttribute('data-theme', 'dark'); } catch (e) {} })();
    function nslToggleTheme() {
        var d = document.documentElement, dark = d.getAttribute('data-theme') === 'dark';
        if (dark) { d.removeAttribute('data-theme'); } else { d.setAttribute('data-theme', 'dark'); }
        try { localStorage.setItem('nsl-theme', dark ? 'light' : 'dark'); } catch (e) {}
        nslThemeLabel();
    }
    function nslThemeLabel() {
        var b = document.getElementById('theme-toggle');
        if (b) b.textContent = document.documentElement.getAttribute('data-theme') === 'dark' ? 'Light' : 'Dark';
    }
    document.addEventListener('DOMContentLoaded', nslThemeLabel);
    </script>
    <link rel="stylesheet" href="styles.css">
    <!-- Expose the API base so client JS (e.g. the scan-status poller) can reach
         the Go API directly (CORS is open). -->
    <script>window.NSL_API = <?= json_encode(API_BASE_URL) ?>;</script>
    <!-- Load synchronously in <head> so nslWatchScan() is defined before any
         inline call in the page body runs. -->
    <script src="scan-status.js"></script>
    <!-- Credential-vault app-bar control (status + unlock/lock). -->
    <script src="vault.js"></script>
</head>
<body class="<?= htmlspecialchars($bodyClass ?? '') ?>">
    <header class="appbar">
        <span class="appbar-brand">NSL-Graph</span>
        <nav class="appbar-nav">
            <?php foreach ($navItems as $file => $label): ?>
                <a href="<?= $file ?>" class="<?= $currentPage === $file ? 'active' : '' ?>"><?= htmlspecialchars($label) ?></a>
            <?php endforeach; ?>
        </nav>
        <div class="appbar-vault" title="Credential vault for stored SSH secrets">
            <span id="vault-status" class="vault-pill">Vault: …</span>
            <button type="button" id="vault-action" onclick="nslVaultAction()" style="display:none">Unlock</button>
            <button type="button" id="vault-lock" onclick="nslVaultLock()" style="display:none">Lock</button>
        </div>
        <form class="appbar-api" method="post">
            <label for="apiUrl">API Base URL:</label>
            <input type="text" id="apiUrl" name="api_base_url" value="<?= htmlspecialchars(API_BASE_URL) ?>">
            <button type="submit">Update</button>
        </form>
        <button type="button" id="theme-toggle" class="theme-toggle" onclick="nslToggleTheme()" title="Toggle dark mode">Dark</button>
    </header>
    <main class="appmain">
