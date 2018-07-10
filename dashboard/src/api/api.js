export function isUnauthorized(e) { 
    return (e.status === 401);
}

export default {
    token: '',

    async decodeResults(result) {
        const decoded = await result.json();
        if (result.status !== 200) {
            let message = decoded.error;
            if (!message) {
                if (result.status === 401) {
                    message = "Unauthorized";
                } else {
                    message = `Unexpected status ${result.status}`;
                }
            }
            throw Object.assign(new Error(message), { status: result.status });
        }
        return decoded;
    },
    
    // apiGet performs a GET request on the API with given local URL.
    // The result is decoded from JSON and returned.
    async get(localURL) {
        let headers = {
            'Accept': 'application/json'
        };
        if (this.token) {
            headers['Authorization'] = `bearer ${this.token}`; 
        }
        const result = await fetch(localURL, {headers});
        return this.decodeResults(result);
    },
    
    // apiPost performs a POST request on the API with given local URL and given data.
    // The result is decoded from JSON and returned.
    async post(localURL, body) {
        let headers = {
            'Accept': 'application/json',
            'Content-Type': 'application/json'
        };
        if (this.token) {
            headers['Authorization'] = `bearer ${this.token}`; 
        }
        const result = await fetch(localURL, {
            method: 'POST',
            headers,
            body: JSON.stringify(body)
        });
        return this.decodeResults(result);
    }
};
