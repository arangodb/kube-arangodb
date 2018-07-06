const PREFIX = "kube-arangodb:v1:";

export function getSessionItem(key) {
  const item = sessionStorage.getItem(`${PREFIX}${key}`);
  if (item) {
    try {
      return JSON.parse(item);
    } catch (e) {}
  }
  return undefined;
}

export function setSessionItem(key, value) {
  sessionStorage.setItem(`${PREFIX}${key}`, JSON.stringify(value));
  return value;
}
