const localPreview = ["localhost", "127.0.0.1", "[::1]"].includes(location.hostname)
  && location.port !== "8080";
const portalAddress = localPreview ? "http://localhost:8080/login" : "/login";
document.getElementById("portalLink").href = portalAddress;
location.replace(portalAddress);
