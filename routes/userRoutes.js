const router = require("express").Router();
const {
  register,
  login,
  updateUser,
} = require("../controllers/userController");
const authMiddleware = require("../middlewares/authMiddleware");

router.post("/register", register);
router.post("/login", login);
router.patch("/", authMiddleware, updateUser);

module.exports = router;
