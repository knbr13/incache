require("dotenv").config();
const express = require("express");
const mongoose = require("mongoose");

const app = express();

const userRoutes = require("./routes/userRoutes");
app.use("/user", userRoutes);

const PORT = process.env.PORT || 8000;

mongoose.connect(process.env.MONGO_URI).then(() => {
  app.listen(PORT, () => {
    console.log("the server is up, and listening on port", PORT);
  });
});
