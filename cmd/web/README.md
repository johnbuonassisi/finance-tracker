# Testing Login

# 1. Register a user
curl -i -X POST \
  -d "username=testuser1" \
  -d "password=password1" \
  "http://localhost:8080/register"

# 2. Login and save cookies
curl -i -c /tmp/finance-tracker-cookies.txt -X POST \
  -d "username=testuser1" \
  -d "password=password1" \
  "http://localhost:8080/login"

# 3. View the saved cookies so you can copy the csrf value if needed
cat /tmp/finance-tracker-cookies.txt

# 4. Call the protected endpoint
# Replace <csrf-token> with the csrf cookie value from the cookie jar
curl -i -b /tmp/finance-tracker-cookies.txt -X POST \
  -H "csrf: <csrf-token>" \
  -d "username=testuser1" \
  "http://localhost:8080/protected"

# 5. Logout
curl -i -b /tmp/finance-tracker-cookies.txt -c /tmp/finance-tracker-cookies.txt -X POST \
  -H "csrf: <csrf-token>" \
  -d "username=testuser1" \
  "http://localhost:8080/logout"

# 6. Try protected again after logout, should fail
curl -i -b /tmp/finance-tracker-cookies.txt -X POST \
  -H "csrf: <csrf-token>" \
  -d "username=testuser1" \
  "http://localhost:8080/protected"
