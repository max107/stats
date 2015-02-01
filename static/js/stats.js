var Stat = function() {
    return this.init();
};

Stat.prototype = {
    fingerprint: null,

    init: function() {
        this.fingerprint = new Fingerprint({
            canvas: true,
            screen_resolution: true
        }).get();

        var data = {
            host: window.location.host,
            screen: {
                availWidth: window.screen.availWidth,
                availHeight: window.screen.availHeight,
                width: window.screen.width,
                height: window.screen.height,
                colorDepth: window.screen.colorDepth,
                pixelDepth: window.screen.pixelDepth,
                orientation: window.screen.orientation ? {
                    angle: window.screen.orientation.angle,
                    type: window.screen.orientation.type
                } : {}
            },
            navigator: {
                appCodeName: window.navigator.appCodeName,
                appName: window.navigator.appName,
                appVersion: window.navigator.appVersion,
                cookieEnabled: window.navigator.cookieEnabled,
                doNotTrack: window.navigator.doNotTrack,
                language: window.navigator.language,
                maxTouchPoints: window.navigator.maxTouchPoints,
                platform: window.navigator.platform,
                product: window.navigator.product,
                productSub: window.navigator.productSub,
                userAgent: window.navigator.userAgent,
                vendor: window.navigator.vendor,
                vendorSub: window.navigator.vendorSub
            },
            location: {
                hash: window.location.hash,
                host: window.location.host,
                hostname: window.location.hostname,
                href: window.location.href,
                origin: window.location.origin,
                pathname: window.location.pathname,
                port: window.location.port,
                protocol: window.location.protocol,
                search: window.location.search
            },
            fingerprint: this.fingerprint
        };
        this.ajax("http://localhost:8080", {
            method: "POST",
            data: JSON.stringify(data),
            async: true,
            headers: {
                'Content-Type': 'application/json'
            },
            success: function(response, xhr) {
                console.log(response);
                // console.log('Got users', JSON.parse(response));
            },
            error: function(status, message, xhr) {
                // console.error('Users API returned', status, message);
                console.log(xhr);
            }
        });
    },
    createXHR: function() {
        var xhr;
        if (window.ActiveXObject) {
            try {
                xhr = new ActiveXObject("Microsoft.XMLHTTP");
            } catch (e) {
                console.log(e.message);
                xhr = null;
            }
        } else {
            xhr = new XMLHttpRequest();
        }

        return xhr;
    },
    ajax: function(url, options) {
        options = options || {};
        options.method = options.method || 'GET';
        options.headers = options.headers || {};
        options.success = options.success || function() {};
        options.error = options.error || function() {};
        options.async = typeof options.async === 'undefined' ? true : options.async;

        // var client = new XMLHttpRequest();
        var client = this.createXHR();
        client.open(options.method, url);
        client.overrideMimeType('application/json');

        for (var i in options.headers) {
            if (options.headers.hasOwnProperty(i)) {
                client.setRequestHeader(i, options.headers[i]);
            }
        }

        client.send(options.data);
        client.onreadystatechange = function() {
            if (this.readyState == 4 && this.status == 200) {
                options.success(this.responseText, this);
            } else if (this.readyState == 4) {
                options.error(this.status, this.statusText, this);
            }
        };

        return client;
    }
};

new Stat;