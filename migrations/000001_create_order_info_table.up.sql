CREATE TABLE order_service.orders(
    id BIGINT PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY ,
    user_id BIGINT NOT NULL REFERENCES sso.users(id),
    item_id BIGINT NOT NULL REFERENCES catalogue.item_info(id)
);
