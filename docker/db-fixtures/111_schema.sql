DROP TABLE IF EXISTS public.stock;
create table stock
(
	sku varchar(255) not null,
	warehouse varchar(255) not null,
	quantity integer,

	constraint stock_pk primary key (sku, warehouse)
);
alter table stock owner to username;
create unique index stock_sku_warehouse_uindex on stock (sku, warehouse);


DROP TABLE IF EXISTS public.transaction;
create table transaction
(
	sku varchar(255),
	warehouse varchar(255),
	quantity integer,
	description text,
	inserted_at timestamp not null default now()
);
alter table transaction owner to username;
create index transaction_inserted_at_index on transaction (inserted_at);
create index transaction_sku_index on transaction (sku);
create index transaction_sku_warehouse_index on transaction (sku, warehouse);
create index transaction_warehouse_index on transaction (warehouse);

