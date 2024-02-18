ALTER TABLE basket ADD purchased BOOLEAN DEFAULT false AFTER id;

CREATE TRIGGER update_item
AFTER
UPDATE
  ON basket FOR EACH ROW 
BEGIN 
    IF (new.purchased = true) THEN
        UPDATE items
        SET times_bought = times_bought + 1, last_purchase_date = CURRENT_TIMESTAMP
        WHERE id = new.item_id;
    END IF;
END;