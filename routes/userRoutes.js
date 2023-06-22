const router = require("express").Router();
const {
  register,
  login,
  updateUser,
} = require("../controllers/userController");

router.post("/register", register);
router.post("/login", login);
router.patch("/", updateUser);

module.exports = router;
