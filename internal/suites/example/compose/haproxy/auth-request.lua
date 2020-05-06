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

local http = require("socket.http")

core.register_action("auth-request", { "http-req" }, function(txn, be, path)
	txn:set_var("txn.auth_response_successful", false)

	-- Check whether the given backend exists.
	if core.backends[be] == nil then
		txn:Alert("Unknown auth-request backend '" .. be .. "'")
		txn:set_var("txn.auth_response_code", 500)
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
		txn:set_var("txn.auth_response_code", 500)
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
	local b, c, h = http.request {
		url = "http://" .. addr .. path,
		headers = headers,
		create = core.tcp,
		-- Disable redirects, because DNS does not work here.
		redirect = false,
		-- We do not check body, so HEAD
		method = "HEAD",
	}

	-- Check whether we received a valid HTTP response.
	if b == nil then
		txn:Warning("Failure in auth-request backend '" .. be .. "': " .. c)
		txn:set_var("txn.auth_response_code", 500)
		return
	end

	txn:set_var("txn.auth_response_code", c)

	-- 2xx: Allow request.
    if 200 <= c and c < 300 then
        if h["remote-user"] then
            txn:set_var("txn.auth_user", h["remote-user"])
        end
        if h["remote-groups"] then
            txn:set_var("txn.auth_groups", h["remote-groups"])
        end
		txn:set_var("txn.auth_response_successful", true)
	-- Don't allow other codes.
	-- Codes with Location: Passthrough location at redirect.
	elseif c == 301 or c == 302 or c == 303 or c == 307 or c == 308 then
		txn:set_var("txn.auth_response_location", h["location"])
	-- 401 / 403: Do nothing, everything else: log.
	elseif c ~= 401 and c ~= 403 then
		txn:Warning("Invalid status code in auth-request backend '" .. be .. "': " .. c)
	end
end, 2)