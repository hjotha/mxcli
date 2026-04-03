# REST Integration

Calling external APIs from microflows -- GET, POST, authentication, and error handling.

## Simple GET

```sql
CREATE MICROFLOW Integration.FetchWebpage ()
RETURNS String AS $Content
BEGIN
  $Content = REST CALL GET 'https://example.com/api/status'
    HEADER Accept = 'application/json'
    TIMEOUT 30
    RETURNS String;
  RETURN $Content;
END;
/
```

## GET with URL Parameters

```sql
CREATE MICROFLOW Integration.SearchProducts (
  $Query: String,
  $Page: Integer
)
RETURNS String AS $Response
BEGIN
  $Response = REST CALL GET 'https://api.example.com/search?q={1}&page={2}' WITH (
    {1} = urlEncode($Query),
    {2} = toString($Page)
  )
    HEADER Accept = 'application/json'
    TIMEOUT 60
    RETURNS String;
  RETURN $Response;
END;
/
```

## POST with JSON Body

```sql
CREATE MICROFLOW Integration.CreateCustomer (
  $Name: String,
  $Email: String
)
RETURNS String AS $Response
BEGIN
  $Response = REST CALL POST 'https://api.example.com/customers'
    HEADER 'Content-Type' = 'application/json'
    BODY '{{"name": "{1}", "email": "{2}"}' WITH (
      {1} = $Name,
      {2} = $Email
    )
    TIMEOUT 30
    RETURNS String;
  RETURN $Response;
END;
/
```

## Basic Authentication

```sql
CREATE MICROFLOW Integration.FetchSecureData (
  $Username: String,
  $Password: String
)
RETURNS String AS $Response
BEGIN
  $Response = REST CALL GET 'https://api.example.com/secure/data'
    HEADER Accept = 'application/json'
    AUTH BASIC $Username PASSWORD $Password
    TIMEOUT 30
    RETURNS String;
  RETURN $Response;
END;
/
```

## Error Handling

Use `ON ERROR WITHOUT ROLLBACK` to catch failures and return a fallback instead of rolling back the transaction:

```sql
CREATE MICROFLOW Integration.SafeAPICall (
  $Url: String
)
RETURNS Boolean AS $Success
BEGIN
  DECLARE $Success Boolean = false;

  $Response = REST CALL GET $Url
    HEADER Accept = 'application/json'
    TIMEOUT 30
    RETURNS String
    ON ERROR WITHOUT ROLLBACK {
      LOG ERROR NODE 'Integration' 'API call failed: ' + $Url;
      RETURN false;
    };

  SET $Success = true;
  RETURN $Success;
END;
/
```

## JSON Structure + Import Mapping

Map a JSON response to Mendix entities using a JSON structure and import mapping.

```sql
-- Step 1: Define the JSON structure
CREATE JSON STRUCTURE Integration.JSON_Pet
  SNIPPET '{"id": 1, "name": "Fido", "status": "available"}';

-- Step 2: Create a non-persistent entity to hold the data
CREATE NON-PERSISTENT ENTITY Integration.PetResponse (
  PetId: Integer,
  Name: String,
  Status: String
);
/

-- Step 3: Create the import mapping
CREATE IMPORT MAPPING Integration.IMM_Pet
  WITH JSON STRUCTURE Integration.JSON_Pet
{
  CREATE Integration.PetResponse {
    PetId = id,
    Name = name,
    Status = status
  }
};

-- Step 4: Use in a microflow with IMPORT FROM MAPPING
-- $PetResponse = IMPORT FROM MAPPING Integration.IMM_Pet($JsonContent);
```

## Export Mapping (Entity → JSON)

Serialize entities back to JSON using an export mapping.

```sql
CREATE EXPORT MAPPING Integration.EMM_Pet
  WITH JSON STRUCTURE Integration.JSON_Pet
{
  Integration.PetResponse {
    id = PetId,
    name = Name,
    status = Status
  }
};

-- Use in a microflow with EXPORT TO MAPPING
-- $JsonOutput = EXPORT TO MAPPING Integration.EMM_Pet($PetResponse);
```

## Nested Mappings with Associations

Map nested JSON objects to multiple entities linked by associations.

```sql
CREATE IMPORT MAPPING Integration.IMM_Order
  WITH JSON STRUCTURE Integration.JSON_Order
{
  CREATE Integration.OrderResponse {
    OrderId = orderId KEY,
    CREATE Integration.OrderResponse_CustomerInfo/Integration.CustomerInfo = customer {
      Email = email,
      Name = name
    },
    CREATE Integration.OrderResponse_OrderItem/Integration.OrderItem = items {
      Sku = sku,
      Quantity = quantity,
      Price = price
    }
  }
};
```
