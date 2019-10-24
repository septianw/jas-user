-- -----------------------------------------------------
-- Table `mydb`.`user`
-- -----------------------------------------------------
-- +migrate Up
CREATE TABLE IF NOT EXISTS `user` (
  `uid` INT NOT NULL AUTO_INCREMENT,
  `uname` VARCHAR(191) NOT NULL,
  `upass` TEXT NOT NULL,
  `deleted` TINYINT(1) NOT NULL DEFAULT 0,
  `contact_contactid` INT NOT NULL,
  PRIMARY KEY (`uid`,`uname`))
ENGINE = InnoDB;

-- +migrate Down
DROP TABLE IF EXISTS `user` ;