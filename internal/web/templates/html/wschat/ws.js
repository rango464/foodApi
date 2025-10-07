var loc = window.location;
    var uri = 'ws://localhost:1323/ws';

    if (loc.protocol === 'https:') {
      uri = 'wss://localhost:1323/ws';
    }
    uri += '//' + loc.host;
    uri += loc.pathname + 'ws';

    ws = new WebSocket(uri)

    ws.onopen = function() {
      console.log('Connected')
    }

    ws.onmessage = function(evt) {
      var out = document.getElementById('output');
      out.innerHTML += evt.data + '<br>';
    }

    setInterval(function() {
      ws.send('Hello, Server!');
    }, 1000);