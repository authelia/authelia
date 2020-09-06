-- The MIT License (MIT)
--
-- Copyright (c) 2018 Tim Düsterhus
--
-- Permission is hereby granted, free of charge, to any person obtaining a copy
-- of this software and associated documentation files (the "Software"), to deal
-- in the Software without restriction, including without limitation the rights
-- to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
-- copies of the Software, and to permit persons to whom the Software is
-- furnished to do so, subject to the following conditions:
--
-- The above copyright notice and this permission notice shall be included in all
-- copies or substantial portions of the Software.
--
-- THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
-- IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
-- FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
-- AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
-- LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
-- OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
-- SOFTWARE.

local http = require("haproxy-lua-http")

function set_var_pre_2_2(txn, var, value)
	return txn:set_var(var, value)
end
function set_var_post_2_2(txn, var, value)
	return txn:set_var(var, value, true)
end

set_var = function(txn, var, value)
	local success = pcall(set_var_post_2_2, txn, var, value)
	if success then
		set_var = set_var_post_2_2
	else
		set_var = set_var_pre_2_2
	end

	return set_var(txn, var, value)
end

function sanitize_header_for_variable(header)
	return header:gsub("[^a-zA-Z0-9]", "_")
end


core.register_action("auth-request", { "http-req" }, function(txn, be, path)
	set_var(txn, "txn.auth_response_successful", false)

	-- Check whether the given backend exists.
	if core.backends[be] == nil then
		txn:Alert("Unknown auth-request backend '" .. be .. "'")
		set_var(txn, "txn.auth_response_code", 500)
		return
	end

	-- Check whether the given backend has servers that
	-- are not `DOWN`.
	local addr = nil
	for name, server in pairs(core.backends[be].servers) do
		local status = server:get_stats()['status']
		if status == "no check" or status:find("UP") == 1 then
			addr = server:get_addr()
			break
		end
	end
	if addr == nil then
		txn:Warning("No servers available for auth-request backend: '" .. be .. "'")
		set_var(txn, "txn.auth_response_code", 500)
		return
	end

	-- Transform table of request headers from haproxy's to
	-- socket.http's format.
	local headers = {}
	for header, values in pairs(txn.http:req_get_headers()) do
		if header ~= 'content-length' then
			for i, v in pairs(values) do
				if headers[header] == nil then
					headers[header] = v
				else
					headers[header] = headers[header] .. ", " .. v
				end
			end
		end
	end

	-- Make request to backend.
	local response, err = http.head {
		url = "http://" .. addr .. path,
		headers = headers,
	}

	-- Check whether we received a valid HTTP response.
	if response == nil then
		txn:Warning("Failure in auth-request backend '" .. be .. "': " .. err)
		set_var(txn, "txn.auth_response_code", 500)
		return
	end

	set_var(txn, "txn.auth_response_code", response.status_code)

	for header, value in response:get_headers(true) do
		set_var(txn, "req.auth_response_header." .. sanitize_header_for_variable(header), value)
	end

	-- 2xx: Allow request.
	if 200 <= response.status_code and response.status_code < 300 then
		set_var(txn, "txn.auth_response_successful", true)
	-- Don't allow other codes.
	-- Codes with Location: Passthrough location at redirect.
	elseif response.status_code == 301 or response.status_code == 302 or response.status_code == 303 or response.status_code == 307 or response.status_code == 308 then
		set_var(txn, "txn.auth_response_location", response:get_header("location", "last"))
	-- 401 / 403: Do nothing, everything else: log.
	elseif response.status_code ~= 401 and response.status_code ~= 403 then
		txn:Warning("Invalid status code in auth-request backend '" .. be .. "': " .. response.status_code)
	end
end, 2)
