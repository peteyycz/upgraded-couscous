TARGET = "https://api.api-ninjas.com/v1/quotes"
API_KEY = os.getenv("API_KEY")

function Handle(path, query)
	print("route hit: " .. path .. " with query " .. query)
end

-- Not implemented yet
Routing = {}
Routing["default"] = function()
	local headers = {}
	headers["X-API-Key"] = API_KEY
	return headers
end
Routing["/"] = function()
	print("123")
end
