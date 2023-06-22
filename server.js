require("dotenv").config();
const express = require("express");

const app = express();

app.listen(process.env.PORT, () => {
    console.log("the server is up, and listening on port", process.env.PORT);
});