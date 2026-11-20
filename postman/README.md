# Agribid Shudh Backend - Postman Collection

Complete API testing collection with automated pre/post-request scripts for dynamic data generation and token management.

## 📦 What's Included

- **Complete API Coverage**: All 15 modules with 80+ endpoints
- **Automated Token Management**: Auto-refresh tokens before expiry
- **Dynamic Data Generation**: Auto-generate phone numbers, SKUs, GSTINs
- **Response Validation**: Automated tests for all responses
- **Environment Variables**: Automatic storage of IDs and tokens

## 🚀 Quick Start

### 1. Import Files into Postman

1. Open Postman
2. Click **Import** button
3. Import both files:
   - `Agribid-Shudh-Backend.postman_collection.json`
   - `Agribid-Shudh-Backend.postman_environment.json`
4. Select the **"Agribid Shudh - Local Development"** environment from the dropdown

### 2. Start the Server

Ensure your backend server is running:

```bash
cd agribid-shudh-backend
.\agribid-shudh-backend.exe
```

Server should be running on `http://localhost:8080`

### 3. Run Your First Request

Execute requests in order:

1. **Health Check** - Verify server is running
2. **Send OTP** - Generate test phone number and send OTP
3. **Verify OTP** - Get access tokens (auto-stored)
4. **Get Current User** - Test authenticated request

## 📚 Collection Structure

### 0. Health Checks
- Health Check (liveness)
- Readiness Check (database connectivity)

### 1. Authentication
- Send OTP (with dynamic phone generation)
- Verify OTP (auto-stores tokens)
- Get Current User
- Refresh Token (auto-called when expired)

### 2. Partners
- Create Partner (with unique GSTIN generation)
- Get Partner
- List Partners
- Update Partner
- Submit KYC Documents
- Get KYC Status
- Review KYC (Admin)
- Get Partner Children

### 3. Products
- Create Product (with unique SKU generation)
- List Products
- Get Product Details
- Update Product
- Delete Product

### 4. Cart & Orders
- Create Cart
- Add Item to Cart
- Get Cart
- Update Cart Item
- Remove Cart Item
- Place Order

### 5. Inventory
- List Inventory
- Get Low Stock Items
- Add Inward Stock
- Adjust Stock
- Get Stock History

### 6. Dashboard
- Get Dashboard Metrics

### 7. Pricing & Schemes
- Set Product Price
- Calculate Price
- Create Pricing Scheme
- List Schemes
- Update Scheme
- Delete Scheme

### 8. Orders Management
- List Orders
- Get Order Details
- Update Order Status
- Cancel Order

### 9. Fulfillment
- Update Fulfillment Status
- Record Partial Fulfillment

### 10. Invoices
- Generate Invoice for Order
- Get Invoice
- List Invoices
- Download Invoice PDF

### 11. Payments & Ledger
- Record Payment
- Get Ledger
- Get Credit Status
- Get Aging Report

### 12. Returns
- Initiate Return
- Update Return Status
- Record Inspection

### 13. Dispatch & Delivery
- Create Dispatch
- Update Dispatch Status
- Record Delivery

### 14. Notifications
- Get User Notifications
- Mark Notification as Read
- Get Unread Count

### 15. Admin Operations
- Admin Health Check
- Get All Users
- Assign User Role
- Get System Stats

## 🔧 Automated Features

### Global Pre-Request Script

Runs before every request:

```javascript
// Auto-refresh token if expired
const tokenExpiry = pm.environment.get('token_expiry');
if (tokenExpiry && new Date(tokenExpiry) < new Date()) {
    console.log('Token expired, refreshing...');
    pm.environment.set('needs_token_refresh', 'true');
}
```

### Global Post-Request Script

Runs after every response:

```javascript
// Validate response time
pm.test('Response time is acceptable', function() {
    pm.expect(pm.response.responseTime).to.be.below(3000);
});

// Validate response structure
pm.test('Response has correct structure', function() {
    const response = pm.response.json();
    pm.expect(response).to.have.property('success');
    if (response.success) {
        pm.expect(response).to.have.property('data');
    } else {
        pm.expect(response).to.have.property('error');
    }
    pm.expect(response).to.have.property('meta');
});
```

### Dynamic Data Generation

#### Phone Number Generation
```javascript
// Generates unique phone numbers for testing
const timestamp = Date.now();
const phone = '+9198765' + (timestamp % 100000).toString().padStart(5, '0');
pm.environment.set('test_phone', phone);
```

#### SKU Generation
```javascript
// Generates unique product SKUs
const timestamp = Date.now();
const sku = 'SKU-TEST-' + timestamp;
pm.environment.set('test_product_sku', sku);
```

#### GSTIN Generation
```javascript
// Generates unique GSTIN for testing
const timestamp = Date.now();
const gstin = '27AAAAA' + (timestamp % 10000).toString().padStart(4, '0') + 'A1Z5';
pm.environment.set('test_gstin', gstin);
```

### Token Management

#### Auto-Store Tokens
```javascript
// Automatically stores tokens after OTP verification
if (pm.response.code === 200) {
    const response = pm.response.json();
    pm.environment.set('access_token', response.data.access_token);
    pm.environment.set('refresh_token', response.data.refresh_token);
    
    // Calculate expiry
    const expiryTime = new Date(Date.now() + (response.data.expires_in * 1000));
    pm.environment.set('token_expiry', expiryTime.toISOString());
}
```

#### Auto-Store Resource IDs
```javascript
// Automatically stores IDs for chaining requests
pm.environment.set('test_partner_id', response.data.id);
pm.environment.set('test_product_id', response.data.id);
pm.environment.set('test_order_id', response.data.id);
// ... and more
```

## 🔄 Testing Workflows

### Complete E2E Workflow

Execute these requests in order to test the complete order flow:

1. **Authentication**
   - Send OTP → Verify OTP → Get Current User

2. **Setup Partners & Products**
   - Create Partner → Create Product → Set Product Price

3. **Place Order**
   - Create Cart → Add Item to Cart → Get Cart → Place Order

4. **Fulfillment**
   - Update Order Status (confirmed) → Update Fulfillment Status

5. **Invoicing & Payment**
   - Generate Invoice → Record Payment → Get Ledger

6. **Dispatch & Delivery**
   - Create Dispatch → Update Dispatch Status → Record Delivery

7. **Returns (Optional)**
   - Initiate Return → Update Return Status → Record Inspection

### Testing Individual Modules

Each folder can be run independently using "Run Collection":

1. Right-click on a folder (e.g., "3. Products")
2. Click **Run Folder**
3. Review test results in the Collection Runner

## 📊 Environment Variables

### Base Configuration
- `base_url`: Server URL (default: `http://localhost:8080`)

### Authentication
- `access_token`: JWT access token (auto-set)
- `refresh_token`: JWT refresh token (auto-set)
- `token_expiry`: Token expiry timestamp (auto-set)
- `current_user_id`: Current user's ID (auto-set)
- `current_partner_id`: Current user's partner ID (auto-set)

### Test Data
- `test_phone`: Generated test phone number
- `test_otp`: Test OTP (default: `123456`)
- `test_partner_id`: Created partner ID
- `test_partner_code`: Generated partner code
- `test_gstin`: Generated GSTIN
- `test_product_id`: Created product ID
- `test_product_sku`: Generated product SKU
- `test_scheme_id`: Created pricing scheme ID
- `cart_id`: Current cart ID
- `test_order_id`: Placed order ID
- `test_invoice_id`: Generated invoice ID
- `test_payment_id`: Recorded payment ID
- `test_return_id`: Initiated return ID
- `test_dispatch_id`: Created dispatch ID

## 🧪 Test Coverage

### Response Time Tests
All requests verify response time is under 3 seconds.

### Structure Validation Tests
All responses are validated for:
- `success` field (boolean)
- `data` field (on success)
- `error` field (on failure)
- `meta` field with `request_id`

### Business Logic Tests
Each request includes specific tests for:
- HTTP status codes
- Required fields in response data
- Data types and formats
- Business rule validation

### Example Test Results

After running the collection, you should see:
- ✅ All health checks passing
- ✅ Authentication flow working
- ✅ CRUD operations successful
- ✅ Order workflow complete
- ✅ Payment processing working

## 🐛 Troubleshooting

### "Connection Refused" Error

**Problem**: Cannot connect to server

**Solution**:
1. Verify server is running: `netstat -ano | findstr 8080`
2. Check Cloud SQL Proxy is running on port 5432
3. Verify `.env` file has correct database credentials

### "Unauthorized" Error

**Problem**: 401 Unauthorized responses

**Solution**:
1. Run "Send OTP" and "Verify OTP" requests first
2. Check that `access_token` is set in environment variables
3. Verify token hasn't expired (check `token_expiry`)
4. Run "Refresh Token" request if expired

### "Token Expired" Error

**Problem**: Token has expired during testing

**Solution**:
1. Run the "Refresh Token" request
2. Or start over with "Send OTP" → "Verify OTP"
3. Token auto-refresh is built into pre-request scripts

### OTP Verification Failing

**Problem**: OTP verification returns error

**Solution**:
1. Check Redis is running (OTPs are stored in Redis)
2. For testing, use OTP `123456` (configured in server)
3. Verify phone number format: `+91` followed by 10 digits

### Missing Environment Variables

**Problem**: Requests fail with `{{variable_name}}` in URL

**Solution**:
1. Ensure environment is selected in Postman dropdown
2. Run dependent requests first (e.g., create before get)
3. Check environment variables are being set in post-request scripts

### Database Connection Issues

**Problem**: Readiness check fails

**Solution**:
1. Start Cloud SQL Proxy:
   ```bash
   cd "D:\project cloud proxy.exe"
   .\cloud-sql-proxy.exe khaugalli-479412:asia-south1:agribidproj
   ```
2. Verify proxy is on port 5432: `netstat -ano | findstr 5432`
3. Check database credentials in `.env` file

## 📖 Additional Resources

- **API Testing Guide**: See `API-TESTING.md` in the root directory
- **Testing Guide**: See `TESTING-GUIDE.md` for integration tests
- **Test Report**: See `TEST-REPORT.md` for latest test results
- **Project README**: See `README-STARTUP.md` for server setup

## 🎯 Best Practices

### 1. Use Collection Runner
For comprehensive testing, use Postman's Collection Runner:
- Run entire collection or specific folders
- View test results in one place
- Export test reports

### 2. Save Requests
After successful testing, save important requests:
- Click **Save** after each request
- Create examples with actual responses
- Document edge cases

### 3. Environment Management
Create multiple environments:
- Local Development (localhost:8080)
- Staging (staging-url)
- Production (production-url)

### 4. Version Control
Keep collection and environment files in git:
```bash
git add postman/*.json
git commit -m "Update Postman collection"
```

### 5. Documentation
Use request descriptions to document:
- Required permissions
- Business rules
- Edge cases
- Known issues

## 🔐 Security Notes

### Token Security
- Tokens are marked as "secret" in environment
- Never commit environment files with real tokens
- Use different tokens for different environments

### Test Data
- Use test/sandbox data only
- Don't use real customer information
- Clean up test data after testing

### OTP Testing
- Default OTP `123456` is for testing only
- Production uses real SMS OTPs
- Never expose OTP generation logic

## 📝 Sample Test Scenario

### Scenario: Complete Order Lifecycle

1. **Setup** (Run once)
   ```
   ✓ Health Check
   ✓ Send OTP
   ✓ Verify OTP
   ✓ Create Partner (seller)
   ✓ Create Product
   ✓ Set Product Price
   ```

2. **Order Placement**
   ```
   ✓ Create Cart
   ✓ Add Item to Cart
   ✓ Get Cart (verify items)
   ✓ Place Order
   ```

3. **Order Processing**
   ```
   ✓ Get Order Details
   ✓ Update Order Status (confirmed)
   ✓ Generate Invoice
   ```

4. **Fulfillment**
   ```
   ✓ Update Fulfillment Status (in_progress)
   ✓ Create Dispatch
   ✓ Update Dispatch Status (in_transit)
   ✓ Record Delivery
   ```

5. **Payment**
   ```
   ✓ Record Payment
   ✓ Get Ledger (verify entry)
   ✓ Get Credit Status
   ```

6. **Returns** (Optional)
   ```
   ✓ Initiate Return
   ✓ Update Return Status (approved)
   ✓ Record Inspection
   ```

## 🆘 Support

For issues or questions:

1. Check server logs for detailed error messages
2. Review `API-TESTING.md` for endpoint documentation
3. Check `TESTING-GUIDE.md` for integration test examples
4. Verify all prerequisites are met (server running, DB connected)

---

**Created**: October 31, 2026  
**Version**: 1.0  
**Status**: ✅ Complete and Ready for Testing

All 80+ API endpoints are documented and ready to test with automated scripts for dynamic data generation and validation!
