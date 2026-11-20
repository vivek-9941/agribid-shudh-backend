# Complete Workflow Testing Guide

This guide demonstrates how to test complete business workflows using the Postman collection with automated scripts.

## 🎯 Testing Philosophy

Our Postman collection uses:
- **Pre-request scripts** to generate dynamic test data
- **Post-request scripts** to validate responses and store IDs
- **Environment variables** to chain requests together
- **Automated tests** to verify business logic

## 📋 Prerequisites

1. ✅ Server running on `http://localhost:8080`
2. ✅ Database connected (Cloud SQL Proxy on port 5432)
3. ✅ Postman collection and environment imported
4. ✅ Environment "Agribid Shudh - Local Development" selected

## 🔄 Complete Workflows

---

## Workflow 1: User Onboarding & Authentication

### Goal
Register a new user, verify phone, and get authenticated.

### Steps

**1.1 Send OTP**
- **Request**: `POST /api/v1/auth/otp/send`
- **Pre-request Script**:
  ```javascript
  // Generates unique phone: +9198765XXXXX
  const timestamp = Date.now();
  const phone = '+9198765' + (timestamp % 100000).toString().padStart(5, '0');
  pm.environment.set('test_phone', phone);
  console.log('Generated test phone:', phone);
  ```
- **Expected**: OTP sent to phone (stored in Redis)
- **Auto-sets**: `test_phone`

**1.2 Verify OTP**
- **Request**: `POST /api/v1/auth/otp/verify`
- **Body**: Uses `{{test_phone}}` and `{{test_otp}}` (123456)
- **Post-request Script**:
  ```javascript
  // Stores tokens and calculates expiry
  pm.environment.set('access_token', response.data.access_token);
  pm.environment.set('refresh_token', response.data.refresh_token);
  const expiryTime = new Date(Date.now() + (response.data.expires_in * 1000));
  pm.environment.set('token_expiry', expiryTime.toISOString());
  ```
- **Expected**: Access token, refresh token
- **Auto-sets**: `access_token`, `refresh_token`, `token_expiry`

**1.3 Get Current User**
- **Request**: `GET /api/v1/auth/me`
- **Headers**: `Authorization: Bearer {{access_token}}`
- **Post-request Script**:
  ```javascript
  // Stores user ID and partner ID
  pm.environment.set('current_user_id', response.data.id);
  if (response.data.partner_id) {
      pm.environment.set('current_partner_id', response.data.partner_id);
  }
  ```
- **Expected**: User profile with phone, roles
- **Auto-sets**: `current_user_id`, `current_partner_id`

### Success Criteria
- ✅ Unique phone generated
- ✅ OTP verified successfully
- ✅ Tokens stored in environment
- ✅ User profile retrieved
- ✅ All tests passing

---

## Workflow 2: Partner Registration & KYC

### Goal
Register a new distributor partner and submit KYC documents.

### Steps

**2.1 Create Partner**
- **Request**: `POST /api/v1/partners`
- **Pre-request Script**:
  ```javascript
  // Generates unique partner code and GSTIN
  const timestamp = Date.now();
  const partnerCode = 'TST-' + timestamp;
  const gstin = '27AAAAA' + (timestamp % 10000).toString().padStart(4, '0') + 'A1Z5';
  pm.environment.set('test_partner_code', partnerCode);
  pm.environment.set('test_gstin', gstin);
  ```
- **Body**:
  ```json
  {
    "type": "distributor",
    "business_name": "Test Distributors Pvt Ltd",
    "trade_name": "Test Dist",
    "gstin": "{{test_gstin}}",
    "pan": "AAAAA1234A",
    "state_code": "27",
    "address": {
      "line1": "123 Test Street",
      "city": "Mumbai",
      "state": "Maharashtra",
      "pincode": "400001"
    }
  }
  ```
- **Post-request Script**:
  ```javascript
  pm.environment.set('test_partner_id', response.data.id);
  ```
- **Expected**: Partner created with unique GSTIN
- **Auto-sets**: `test_partner_id`, `test_partner_code`, `test_gstin`

**2.2 Submit KYC Documents**
- **Request**: `POST /api/v1/partners/{{test_partner_id}}/kyc`
- **Body**:
  ```json
  {
    "documents": [
      {
        "doc_type": "gst_certificate",
        "file_url": "https://storage.example.com/docs/gst.pdf"
      },
      {
        "doc_type": "pan_card",
        "file_url": "https://storage.example.com/docs/pan.pdf"
      }
    ]
  }
  ```
- **Expected**: KYC status "pending_verification"

**2.3 Get KYC Status**
- **Request**: `GET /api/v1/partners/{{test_partner_id}}/kyc`
- **Expected**: List of submitted documents with status

**2.4 Review KYC (Admin)**
- **Request**: `PATCH /api/v1/partners/{{test_partner_id}}/kyc`
- **Body**:
  ```json
  {
    "status": "approved",
    "notes": "All documents verified"
  }
  ```
- **Expected**: KYC status changed to "approved"

### Success Criteria
- ✅ Partner created with unique GSTIN
- ✅ KYC documents submitted
- ✅ KYC status retrievable
- ✅ Admin can approve KYC
- ✅ Partner status changes to "active"

---

## Workflow 3: Product Catalog Setup

### Goal
Create products, set prices, and configure schemes.

### Steps

**3.1 Create Product**
- **Request**: `POST /api/v1/products`
- **Pre-request Script**:
  ```javascript
  // Generates unique SKU
  const timestamp = Date.now();
  const sku = 'SKU-TEST-' + timestamp;
  pm.environment.set('test_product_sku', sku);
  ```
- **Body**:
  ```json
  {
    "sku": "{{test_product_sku}}",
    "name": "Test Product - Cooking Oil 1L",
    "description": "Test product for API testing",
    "brand": "TestBrand",
    "hsn_code": "15071000",
    "uom": "LTR",
    "weight_grams": 1000,
    "status": "active"
  }
  ```
- **Post-request Script**:
  ```javascript
  pm.environment.set('test_product_id', response.data.id);
  ```
- **Expected**: Product created with unique SKU
- **Auto-sets**: `test_product_id`, `test_product_sku`

**3.2 Set Product Price**
- **Request**: `POST /api/v1/products/{{test_product_id}}/prices`
- **Body**:
  ```json
  {
    "role_code": "DISTRIBUTOR",
    "mrp": "150.00",
    "base_price": "120.00",
    "effective_from": "2026-10-01T00:00:00Z",
    "effective_to": "2026-12-31T23:59:59Z"
  }
  ```
- **Expected**: Price set for DISTRIBUTOR role

**3.3 Calculate Price**
- **Request**: `POST /api/v1/products/{{test_product_id}}/calculate-price`
- **Body**:
  ```json
  {
    "quantity": 10
  }
  ```
- **Expected**: Calculated price with applicable schemes
- **Tests**:
  ```javascript
  pm.test('Price calculation correct', function() {
      pm.expect(response.data.base_price).to.eql(120.00);
      pm.expect(response.data.quantity).to.eql(10);
  });
  ```

**3.4 Create Pricing Scheme**
- **Request**: `POST /api/v1/schemes`
- **Pre-request Script**:
  ```javascript
  const timestamp = Date.now();
  const schemeName = 'Test Scheme ' + timestamp;
  pm.environment.set('test_scheme_name', schemeName);
  ```
- **Body**:
  ```json
  {
    "name": "{{test_scheme_name}}",
    "type": "percentage",
    "discount_value": "10.00",
    "valid_from": "2026-10-01T00:00:00Z",
    "valid_to": "2026-12-31T23:59:59Z",
    "applicability": [
      {
        "target_type": "product",
        "target_id": "{{test_product_id}}"
      }
    ]
  }
  ```
- **Post-request Script**:
  ```javascript
  pm.environment.set('test_scheme_id', response.data.id);
  ```
- **Expected**: 10% discount scheme created
- **Auto-sets**: `test_scheme_id`

### Success Criteria
- ✅ Product created with unique SKU
- ✅ Base price set for role
- ✅ Price calculation working
- ✅ Discount scheme applied
- ✅ Final price reflects discount

---

## Workflow 4: Complete Order Lifecycle

### Goal
Place order, fulfill, invoice, and deliver.

### Steps

**4.1 Create Cart**
- **Request**: `POST /api/v1/cart`
- **Body**:
  ```json
  {
    "seller_id": "{{test_partner_id}}"
  }
  ```
- **Post-request Script**:
  ```javascript
  pm.environment.set('cart_id', response.data.id);
  ```
- **Expected**: Empty cart created
- **Auto-sets**: `cart_id`

**4.2 Add Item to Cart**
- **Request**: `POST /api/v1/cart/items`
- **Body**:
  ```json
  {
    "product_id": "{{test_product_id}}",
    "quantity": 10
  }
  ```
- **Expected**: Item added with calculated price
- **Tests**:
  ```javascript
  pm.test('Item added to cart', function() {
      pm.expect(response.data.quantity).to.eql(10);
      pm.expect(response.data.product_id).to.eql(pm.environment.get('test_product_id'));
  });
  ```

**4.3 Get Cart**
- **Request**: `GET /api/v1/cart`
- **Expected**: Cart with 1 item, subtotal calculated
- **Tests**:
  ```javascript
  pm.test('Cart has items', function() {
      pm.expect(response.data.items).to.have.lengthOf(1);
      pm.expect(response.data.subtotal).to.be.above(0);
  });
  ```

**4.4 Place Order**
- **Request**: `POST /api/v1/orders`
- **Body**:
  ```json
  {
    "delivery_address": {
      "line1": "456 Delivery Street",
      "city": "Pune",
      "state": "Maharashtra",
      "pincode": "411001"
    },
    "notes": "Test order from API collection"
  }
  ```
- **Post-request Script**:
  ```javascript
  pm.environment.set('test_order_id', response.data.id);
  console.log('Order placed:', response.data.order_number);
  ```
- **Expected**: Order created with order_number
- **Auto-sets**: `test_order_id`

**4.5 Update Order Status**
- **Request**: `PATCH /api/v1/orders/{{test_order_id}}/status`
- **Body**:
  ```json
  {
    "status": "confirmed"
  }
  ```
- **Expected**: Order status changed to "confirmed"

**4.6 Generate Invoice**
- **Request**: `POST /api/v1/invoices/order/{{test_order_id}}`
- **Post-request Script**:
  ```javascript
  pm.environment.set('test_invoice_id', response.data.id);
  console.log('Invoice generated:', response.data.invoice_number);
  ```
- **Expected**: Invoice with GST calculations
- **Auto-sets**: `test_invoice_id`
- **Tests**:
  ```javascript
  pm.test('Invoice has GST breakdown', function() {
      pm.expect(response.data).to.have.property('subtotal');
      pm.expect(response.data).to.have.property('total_tax');
      pm.expect(response.data).to.have.property('grand_total');
      // Verify: grand_total = subtotal + total_tax
      const calculated = parseFloat(response.data.subtotal) + parseFloat(response.data.total_tax);
      pm.expect(parseFloat(response.data.grand_total)).to.eql(calculated);
  });
  ```

**4.7 Update Fulfillment Status**
- **Request**: `PATCH /api/v1/fulfillment/orders/{{test_order_id}}/status`
- **Body**:
  ```json
  {
    "status": "in_progress",
    "notes": "Starting fulfillment process"
  }
  ```
- **Expected**: Fulfillment status updated

**4.8 Create Dispatch**
- **Request**: `POST /api/v1/dispatch`
- **Body**:
  ```json
  {
    "order_id": "{{test_order_id}}",
    "vehicle_number": "MH12AB1234",
    "driver_name": "Rajesh Kumar",
    "driver_phone": "+919876543210",
    "expected_delivery_date": "2026-11-05T18:00:00Z"
  }
  ```
- **Post-request Script**:
  ```javascript
  pm.environment.set('test_dispatch_id', response.data.id);
  ```
- **Expected**: Dispatch created
- **Auto-sets**: `test_dispatch_id`

**4.9 Update Dispatch Status**
- **Request**: `PATCH /api/v1/dispatch/{{test_dispatch_id}}/status`
- **Body**:
  ```json
  {
    "status": "in_transit",
    "notes": "Vehicle left warehouse at 10 AM"
  }
  ```
- **Expected**: Dispatch in transit

**4.10 Record Delivery**
- **Request**: `POST /api/v1/dispatch/{{test_dispatch_id}}/delivery`
- **Body**:
  ```json
  {
    "delivered_at": "2026-11-05T15:30:00Z",
    "received_by": "Store Manager",
    "signature_url": "https://storage.example.com/signatures/sig123.png",
    "notes": "Delivered successfully"
  }
  ```
- **Expected**: Delivery recorded, order status changes to "delivered"

**4.11 Record Payment**
- **Request**: `POST /api/v1/payments`
- **Body**:
  ```json
  {
    "seller_id": "{{test_partner_id}}",
    "amount": "10000.00",
    "method": "bank_transfer",
    "reference_number": "TXN1234567890",
    "invoice_id": "{{test_invoice_id}}",
    "notes": "Payment via NEFT"
  }
  ```
- **Post-request Script**:
  ```javascript
  pm.environment.set('test_payment_id', response.data.id);
  ```
- **Expected**: Payment recorded
- **Auto-sets**: `test_payment_id`

**4.12 Verify Ledger**
- **Request**: `GET /api/v1/ledger`
- **Expected**: Ledger entry for payment
- **Tests**:
  ```javascript
  pm.test('Payment recorded in ledger', function() {
      const ledgerEntries = response.data;
      const paymentEntry = ledgerEntries.find(e => e.payment_id === pm.environment.get('test_payment_id'));
      pm.expect(paymentEntry).to.not.be.undefined;
  });
  ```

### Success Criteria
- ✅ Cart created and items added
- ✅ Order placed successfully
- ✅ Invoice generated with correct GST
- ✅ Fulfillment tracked
- ✅ Dispatch created and tracked
- ✅ Delivery recorded
- ✅ Payment recorded
- ✅ Ledger updated
- ✅ All status transitions valid

---

## Workflow 5: Returns & Refunds

### Goal
Initiate return, inspect items, and process refund.

### Steps

**5.1 Initiate Return**
- **Request**: `POST /api/v1/returns`
- **Body**:
  ```json
  {
    "order_id": "{{test_order_id}}",
    "items": [
      {
        "product_id": "{{test_product_id}}",
        "quantity": 2,
        "reason": "damaged"
      }
    ],
    "notes": "Products arrived damaged"
  }
  ```
- **Post-request Script**:
  ```javascript
  pm.environment.set('test_return_id', response.data.id);
  console.log('Return initiated:', response.data.return_number);
  ```
- **Expected**: Return created with return_number
- **Auto-sets**: `test_return_id`

**5.2 Update Return Status (Approve)**
- **Request**: `PATCH /api/v1/returns/{{test_return_id}}/status`
- **Body**:
  ```json
  {
    "status": "approved",
    "notes": "Return approved by admin"
  }
  ```
- **Expected**: Return status "approved"

**5.3 Record Inspection**
- **Request**: `POST /api/v1/returns/{{test_return_id}}/inspection`
- **Body**:
  ```json
  {
    "inspection_result": "approved",
    "notes": "Products verified as damaged",
    "refund_amount": "240.00"
  }
  ```
- **Expected**: Inspection recorded, refund amount set
- **Tests**:
  ```javascript
  pm.test('Refund amount calculated correctly', function() {
      // 2 items × 120 base price = 240
      pm.expect(response.data.refund_amount).to.eql('240.00');
  });
  ```

**5.4 Verify Credit Note**
- **Request**: `GET /api/v1/ledger`
- **Expected**: Credit note entry in ledger
- **Tests**:
  ```javascript
  pm.test('Credit note in ledger', function() {
      const creditNote = response.data.find(e => 
          e.return_id === pm.environment.get('test_return_id') && 
          e.entry_type === 'credit'
      );
      pm.expect(creditNote).to.not.be.undefined;
      pm.expect(creditNote.amount).to.eql('240.00');
  });
  ```

### Success Criteria
- ✅ Return initiated with reason
- ✅ Return approved by admin
- ✅ Inspection recorded
- ✅ Refund amount calculated
- ✅ Credit note created in ledger
- ✅ Available credit updated

---

## Workflow 6: Inventory Management

### Goal
Track inventory movements, handle low stock.

### Steps

**6.1 Check Initial Inventory**
- **Request**: `GET /api/v1/inventory?product_id={{test_product_id}}`
- **Expected**: Current stock levels

**6.2 Add Inward Stock**
- **Request**: `POST /api/v1/inventory/inward`
- **Body**:
  ```json
  {
    "warehouse_id": "WAREHOUSE_UUID",
    "product_id": "{{test_product_id}}",
    "quantity": 100,
    "reason": "Stock received from manufacturer"
  }
  ```
- **Expected**: Stock increased by 100

**6.3 Place Order (Stock Reservation)**
- Follow "Workflow 4" steps 4.1-4.4
- **Tests**:
  ```javascript
  pm.test('Stock reserved after order', function() {
      // Make another request to get inventory
      pm.sendRequest({
          url: pm.environment.get('base_url') + '/api/v1/inventory',
          method: 'GET',
          header: {
              'Authorization': 'Bearer ' + pm.environment.get('access_token')
          }
      }, function(err, res) {
          const inventory = res.json().data.find(i => i.product_id === pm.environment.get('test_product_id'));
          pm.expect(inventory.reserved_quantity).to.be.above(0);
      });
  });
  ```

**6.4 Adjust Stock (Damage)**
- **Request**: `POST /api/v1/inventory/adjust`
- **Body**:
  ```json
  {
    "warehouse_id": "WAREHOUSE_UUID",
    "product_id": "{{test_product_id}}",
    "quantity": -5,
    "reason": "Damaged goods"
  }
  ```
- **Expected**: Stock decreased by 5

**6.5 Get Low Stock Items**
- **Request**: `GET /api/v1/inventory/low-stock`
- **Expected**: List of products below reorder level

### Success Criteria
- ✅ Inward stock recorded
- ✅ Stock reserved on order
- ✅ Stock released on order cancel
- ✅ Adjustments tracked
- ✅ Low stock alerts working

---

## 🧪 Advanced Testing Scenarios

### Scenario 1: Token Refresh Flow

**Test automatic token refresh:**

1. Get current token expiry:
   ```javascript
   console.log('Token expires:', pm.environment.get('token_expiry'));
   ```

2. Manually expire token:
   ```javascript
   pm.environment.set('token_expiry', new Date('2020-01-01').toISOString());
   ```

3. Make any authenticated request
   - Global pre-request script detects expiry
   - Sets `needs_token_refresh` flag

4. Run "Refresh Token" request
   - New tokens stored
   - Expiry updated

5. Retry original request
   - Should succeed with new token

### Scenario 2: Credit Limit Validation

**Test order rejection when credit limit exceeded:**

1. Get current credit status:
   ```
   GET /api/v1/credit
   ```

2. Note `available_credit` value

3. Try to place order exceeding credit:
   - Create cart with high-value items
   - Attempt to place order

4. Verify error response:
   ```javascript
   pm.test('Order rejected - credit limit exceeded', function() {
       pm.expect(pm.response.code).to.eql(400);
       pm.expect(response.error.code).to.eql('CREDIT_LIMIT_EXCEEDED');
   });
   ```

### Scenario 3: Inventory Insufficient Stock

**Test order rejection when stock unavailable:**

1. Get product inventory:
   ```
   GET /api/v1/inventory?product_id={{test_product_id}}
   ```

2. Note `available_quantity`

3. Try to order more than available:
   ```json
   {
     "product_id": "{{test_product_id}}",
     "quantity": 99999
   }
   ```

4. Verify error:
   ```javascript
   pm.test('Order rejected - insufficient stock', function() {
       pm.expect(response.error.code).to.eql('INSUFFICIENT_STOCK');
   });
   ```

### Scenario 4: Concurrent Orders

**Test race conditions:**

1. Create cart A and cart B for same user
2. Add same product to both carts
3. Place both orders simultaneously
4. Verify:
   - Both succeed (if stock available)
   - Or one fails with insufficient stock
   - Stock reserved correctly

---

## 📊 Collection Runner Usage

### Running Complete Workflow

1. Click **Runner** button in Postman
2. Select collection: "Agribid Shudh Backend API"
3. Select environment: "Agribid Shudh - Local Development"
4. Configure run:
   - **Iterations**: 1
   - **Delay**: 500ms (between requests)
   - **Data file**: (optional CSV for bulk testing)
5. Click **Run**

### Viewing Results

After run completes:
- ✅ Green: All tests passed
- ❌ Red: Test failed (click to see details)
- **Total requests**: Should match expected count
- **Total tests**: All should pass
- **Duration**: Typical run ~30-60 seconds

### Exporting Report

1. After run, click **Export Results**
2. Save as JSON for documentation
3. Share with team for review

---

## 🐛 Debugging Failed Tests

### Common Issues

**Issue**: `{{variable_name}}` appears in URL
- **Cause**: Variable not set in environment
- **Fix**: Run dependent request first (e.g., create before get)

**Issue**: 401 Unauthorized
- **Cause**: Token expired or not set
- **Fix**: Run "Send OTP" → "Verify OTP" first

**Issue**: 404 Not Found
- **Cause**: Resource ID invalid
- **Fix**: Check environment variable has correct UUID

**Issue**: 400 Validation Error
- **Cause**: Request body invalid
- **Fix**: Check required fields and data types

### Debug Mode

Enable Postman Console:
- View → Show Postman Console
- See all requests, responses, and console.log outputs
- Check pre-request and post-request script execution

---

## 📈 Performance Testing

Use Collection Runner with iterations:

```
Iterations: 100
Delay: 100ms
Save responses: false
```

Monitor:
- Average response time
- Failed requests
- Server resource usage

---

## ✅ Quality Checklist

Before marking workflow complete:

- [ ] All requests execute successfully
- [ ] All tests pass (green checkmarks)
- [ ] Environment variables populated correctly
- [ ] Response data validates business rules
- [ ] Error scenarios tested
- [ ] Edge cases covered
- [ ] Performance acceptable (<2s per request)
- [ ] Documentation updated

---

**Created**: October 31, 2026  
**Version**: 1.0  
**Status**: ✅ Complete Workflow Testing Guide

Use these workflows to systematically test all API functionality with confidence!
