const elements = {
	overall: document.querySelector("#overall-status"),
	api: document.querySelector("#api-status"),
	network: document.querySelector("#network-status"),
	checked: document.querySelector("#last-checked"),
	message: document.querySelector("#status-message"),
	refresh: document.querySelector("#refresh-button"),
};

function setBadge(state, label) {
	elements.overall.className = `badge badge-${state}`;
	elements.overall.textContent = label;
}

async function checkStatus() {
	elements.refresh.disabled = true;
	setBadge("pending", "Sjekker");

	try {
		const response = await fetch("/api/status", { cache: "no-store" });
		if (!response.ok) throw new Error(`The API returned HTTP ${response.status}`);

		const status = await response.json();
		const networkOk = status.network === "ok";
		setBadge(networkOk ? "ok" : "error", networkOk ? "I orden" : "Fant feil");
		elements.api.textContent = status.api || "OK";
		elements.network.textContent = networkOk ? "Tilkoblet" : "Utilgjengelig";
		elements.message.textContent =
			status.message || "Backend er tilgjengelig gjennom Kubernetes-tjenesten.";
	} catch (error) {
		setBadge("error", "Utilgjengelig");
		elements.api.textContent = "Utilgjengelig";
		elements.network.textContent = "Ikke sjekket";
		elements.message.textContent = error.message;
	} finally {
		elements.checked.textContent = new Intl.DateTimeFormat(undefined, {
			hour: "2-digit",
			minute: "2-digit",
			second: "2-digit",
		}).format(new Date());
		elements.refresh.disabled = false;
	}
}

elements.refresh.addEventListener("click", checkStatus);
checkStatus();
setInterval(checkStatus, 15000);

console.log("Are, Daniel og Ludvig hilser. Du fant consolen. Prøver du kubectl get pods -A?");
