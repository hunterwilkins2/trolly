CREATE TABLE items (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    price float NOT NULL,
    times_purchased INTEGER NOT NULL,
    in_cart BOOLEAN NOT NULL,
    purchased BOOLEAN NOT NULL,
    last_added DATETIME NOT NULL,
    uid INTEGER NOT NULL,
    FOREIGN KEY (uid) REFERENCES users(id) ON DELETE CASCADE
);

