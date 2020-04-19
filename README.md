# appaka warehouse

a simple warehouse stock manager


## todo

1. getting history data
1. updating stock in batch mode (different quantities in different warehouses in one call), useful for add new products, remove stock because of a new purchase...
1. using Redis for caching

## documentation

### add stock

**POST /api/stock**
```json
{
	"description": "new product!",
	"sku": "P123",
	"warehouse": "W3",
	"quantity": 5
}
```

### remove stock

**POST /api/stock**
```json
{
	"description": "purchase #9876",
	"sku": "P123",
	"warehouse": "W3",
	"quantity": -3
}
```


### get stock

**GET /api/stock**
```json
{
	"sku": "P123"
}
```

returns:
```json
{
    "success": true,
    "message": "Stock for P123@",
    "sku": "P123",
    "data": {
        "W1": 1,
        "W2": 15,
        "W3": 7
    }
}
```

### get stock filtering by warehouse

**GET /api/stock**
```json
{
	"sku": "P123",
	"warehouse": "W3"
}
```

returns:

```json
{
    "success": true,
    "message": "Stock for P123@W3",
    "sku": "P123",
    "data": {
        "W3": 7
    }
}
```

