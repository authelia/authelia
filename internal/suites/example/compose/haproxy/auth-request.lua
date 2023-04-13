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
--
-- SPDX-License-Identifier: MIT

local http = require("haproxy-lua-http")

core.register_action("auth-request", { "http-req" }, function(txn, be, path)
	auth_request(txn, be, path, "HEAD", ".*", "-", "-")
end, 2)

core.register_action("auth-intercept", { "http-req" }, function(txn, be, path, method, hdr_req, hdr_succeed, hdr_fail)
	hdr_req = globToLuaPattern(hdr_req)
	hdr_succeed = globToLuaPattern(hdr_succeed)
	hdr_fail = globToLuaPattern(hdr_fail)
	auth_request(txn, be, path, method, hdr_req, hdr_succeed, hdr_fail)
end, 6)

function globToLuaPattern(glob)
	if glob == "-" then
		return "-"
	end
	-- magic chars: '^', '$', '(', ')', '%', '.', '[', ']', '*', '+', '-', '?'
	-- https://www.lua.org/manual/5.4/manual.html#6.4.1
	--
	-- this chain is:
	-- 1. escaping all the magic chars, adding a `%` in front of all of them,
	--    except the chars being processed later in the chain;
	-- 1.1. all the chars inside the [set] are magic chars and have special
	--      meaning inside a set, so we're also escaping all of them to avoid
	--      misbehavior;
	-- 2. converting "match all" `*` and "match one" `?` to their Lua pattern
	--    counterparts;
	-- 3. adding start and finish boundaries outside the whole string and,
	--    being a comma-separated list, between every single item as well.
	return "^" .. glob:gsub("[%^%$%(%)%%%.%[%]%+%-]", "%%%1"):gsub("*", ".*"):gsub("?", "."):gsub(",", "$,^") .. "$"
end

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

-- header_match checks whether the provided header matches the pattern.
-- pattern is a comma-separated list of Lua Patterns.
function header_match(header, pattern)
	if header == "content-length" or header == "host" or pattern == "-" then
		return false
	end
	for p in pattern:gmatch("[^,]*") do
		if header:match(p) then
			return true
		end
	end
	return false
end

-- Terminates the transaction and sends the provided response to the client.
-- hdr_fail filters header names that should be provided using Lua Patterns.
function send_response(txn, response, hdr_fail)
	local reply = txn:reply()
	if response then
		reply:set_status(response.status_code)
		for header, value in response:get_headers(false) do
			if header_match(header, hdr_fail) then
				reply:add_header(header, value)
			end
		end
		if response.content then
			reply:set_body(response.content)
		end
	else
		reply:set_status(500)
	end
	txn:done(reply)
end

-- auth_request makes the request to the external authentication service
-- and waits for the response. hdr_* params receive a comma-separated
-- list of Lua Patterns used to identify the headers that should be
-- copied between the requests and responses. A dash `-` in these params
-- mean that the headers shouldn't be copied at all.
-- Special values and behavior:
-- * method == "*": call the auth service using the same method used by the client.
-- * hdr_fail == "-": make the Lua script to not terminate the request.
function auth_request(txn, be, path, method, hdr_req, hdr_succeed, hdr_fail)
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
		if header_match(header, hdr_req) then
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
	if method == "*" then
		method = txn.sf:method()
	end
	local response, err = http.send(method:upper(), {
		url = "http://" .. addr .. path,
		headers = headers,
	})

	-- `terminate_on_failure == true` means that the Lua script should send the response
	-- and terminate the transaction in the case of a failure. This will happen when
	-- hdr_fail content isn't a dash `-`.
	local terminate_on_failure = hdr_fail ~= "-"

	-- Check whether we received a valid HTTP response.
	if response == nil then
		txn:Warning("Failure in auth-request backend '" .. be .. "': " .. err)
		set_var(txn, "txn.auth_response_code", 500)
		if terminate_on_failure then
			send_response(txn)
		end
		return
	end

	set_var(txn, "txn.auth_response_code", response.status_code)
	local response_ok = 200 <= response.status_code and response.status_code < 300

	for header, value in response:get_headers(true) do
		set_var(txn, "req.auth_response_header." .. sanitize_header_for_variable(header), value)
		if response_ok and hdr_succeed ~= "-" and header_match(header, hdr_succeed) then
			txn.http:req_set_header(header, value)
		end
	end

	-- response_ok means 2xx: allow request.
	if response_ok then
		set_var(txn, "txn.auth_response_successful", true)
	-- Don't allow codes < 200 or >= 300.
	-- Forward the response to the client if required.
	elseif terminate_on_failure then
		send_response(txn, response, hdr_fail)
	-- Codes with Location: Passthrough location at redirect.
	elseif response.status_code == 301 or response.status_code == 302 or response.status_code == 303 or response.status_code == 307 or response.status_code == 308 then
		set_var(txn, "txn.auth_response_location", response:get_header("location", "last"))
	-- 401 / 403: Do nothing, everything else: log.
	elseif response.status_code ~= 401 and response.status_code ~= 403 then
		txn:Warning("Invalid status code in auth-request backend '" .. be .. "': " .. response.status_code)
	end
end
