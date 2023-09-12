<?php

echo "Example Open Redirect vulnerable page";

header("Location: " . $_GET["url"]);

?>
