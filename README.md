# appaka warehouse

a simple warehouse stock manager


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

### batch stock update

**POST /api/stock/batch**
```json
{
	"key": "order#9876",
    "data": {
      "P123": {
        "W1": -3,
        "W2": -1
      },
      "P99": {
        "W1": -1
      }
    },
}
```

returns:
```json
{
    "success": true,
    "key": "order#9876",
    "data": {
        "P123": {
            "W1": 123,
            "W2": 99
        },
        "P333": {
            "W1": 5
        }
    }
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

## todo

1. getting history data
1. create Dockerfile
1. using Redis for caching
1. migrating to gin-gonic? (https://www.youtube.com/watch?v=LOn1GUsjOF4)
1. new method "check" (/api/stock/check) to know if there is stock




