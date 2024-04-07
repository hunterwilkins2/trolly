CREATE TABLE IF NOT EXISTS basket (
  id int NOT NULL AUTO_INCREMENT,
  user_id varchar(36) NOT NULL,
  item_id int NOT NULL,
  PRIMARY KEY (id),
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (item_id) REFERENCES items(id)
);