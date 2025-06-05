DROP TABLE IF EXISTS historyValue;
CREATE TABLE historyValue (
    history_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    bank_account_id INT NOT NULL,
    valuation FLOAT NOT NULL,
    date_valuation DATE NOT NULL,

    PRIMARY KEY (`history_id`),
    FOREIGN KEY (`bank_account_id`) REFERENCES bankAccount(`account_id`)
);
