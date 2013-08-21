Hare
====
Hare is a HTTP to RabbitMQ publishing gateway, written in Go.
Hare solves some of the issues related to short-lived processes (such as PHP) publishing to RabbitMQ, namely overwhelming RabbitMQ with connections.

Hare connects to your RabbitMQ administration API to retrieve exchange definitions, preventing the need to maintain somewhat irrelevant implementation details in your publishing application logic.


Requirements
------------
 * github.com/bitly/go-simplejson
 * github.com/streadway/amqp

Usage
-----
After adding your details to the configuration constants, start hare with `go run hare.go`.

API
---
To access Hare, use the following API endpoints:

 * Management
<table>
  <tr>
    <td>GET: http://harehost:port/alive</td>
    <td>Responds with an HTTP 200 and a body of Alive, for use in a health check.</td>
  </tr>
  <tr>
    <td>GET: http://harehost:port/stats</td>
    <td>Responds with delivery statistics, and an overview of the current definitions object.</td>
  </tr>
  <tr>
    <td>GET: http://harehost:port/reload</td>
    <td>Reloads the definition object by performing another request to the RabbitMQ administration API. Responds with an overview of the new definitions object.</td>
  </tr>
  <tr>
    <td>GET: http://harehost:port/exit</td>
    <td>Empty response. Stops the Hare gateway, logging the initiator of the request to stdout.</td>
  </tr>
</table>

 * Publishing
<table>
  <tr>
    <td>POST: http://harehost:port/publish/{exchange}/{routingKey} body=message</td>
    <td>Requires POST field of body.</td>
    <td>On success, HTTP 200.</td>
    <td>On failure, can respond with 404 (exchange not found), 400 (no post), 500 (publishing failure)</td>
  </tr>
</table>

Todo
----
 * Add support for more than one vhost.
 * Purge map of definitions on reload requests.
 * Allow for command line args to define port.
 * Pretty up stats a bit.

License
-------
Copyright (c) 2013, Weebly
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:
 * Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.
 * Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.
 * Neither the name of Weebly nor the names of its contributors may be used to endorse or promote products derived from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL <COPYRIGHT HOLDER> BE LIABLE FOR ANY
DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
