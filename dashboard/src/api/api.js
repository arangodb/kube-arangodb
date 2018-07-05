// apiGet performs a GET request on the API with given local URL.
// The result is decoded from JSON and returned.
export async function apiGet(localURL) {
    const result = await fetch(localURL);
    const decoded = await result.json();
    return decoded;
}

