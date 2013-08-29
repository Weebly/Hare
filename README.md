Hare
====
Hare is a HTTP to RabbitMQ publishing gateway, written in Go.
Hare solves some of the issues related to short-lived processes (such as PHP) publishing to RabbitMQ, namely overwhelming RabbitMQ with connections.

Hare is API compatible with https://github.com/gmr/statelessd for publishing.

Requirements
------------
 * github.com/bitly/go-simplejson
 * github.com/streadway/amqp

Usage
-----
 * After performing a `go get` to retrieve the source, edit the configuration.go source to set your application's default settings.
 * `go install github.com/{user}/hare` to compile and install the hare binary in your $GOPATH/bin
 * `./hare` to start.

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
    <td>Responds with simple delivery statistics.</td>
  </tr>
  <tr>
    <td>GET: http://harehost:port/exit</td>
    <td>Empty response. Stops the Hare gateway, logging the initiator of the request to stdout.</td>
  </tr>
</table>

 * Publishing
<table>
  <tr>
    <td>POST:<br />http://user:pass@host:port/{vhost}/{exchange}/{routingKey}</td>
    <td>Requires POST field of body.</td>
    <td>On success, responds HTTP 204.</td>
    <td>On failure, can respond with 403 (bad username/password for given vhost), 401 (no username/password given for vhost), 400 (missing post body), 500 (publishing error)</td>
  </tr>
</table>

Todo
----
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
