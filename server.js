require("dotenv").config();
const express = require("express");

const app = express();

const PORT = process.env.PORT || 8000;
app.listen(PORT, () => {
    console.log("the server is up, and listening on port", PORT);
});