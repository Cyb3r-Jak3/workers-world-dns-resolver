import { Container, getRandom } from "@cloudflare/containers";
import { Hono } from "hono";

export class RESOLVER extends Container {
	// Configure default port for the container
	defaultPort = 8080;
	sleepAfter = "10m";
	enableInternet: boolean = true;
}

type Bindings = {
	RESOLVER: DurableObjectNamespace<RESOLVER>;
};

export interface DNSServerResponse {
	server: string;
	values: string[];
	server_address: string;
	ttl: number;
	duration: number; // nanoseconds, matches Go's time.Duration
	duration_string: string;
}

// Represents the overall lookup response
export interface LookupResponse {
	question: string;
	type: string;
	answers: DNSServerResponse[];
	location: string;
	region: string;
	country: string;
	total_duration: number; // nanoseconds, matches Go's time.Duration
	total_duration_string: string;
}

const app = new Hono<{ Bindings: Bindings }>().basePath("/api/v1");

app.get("/lookup", async (c) => {
	const { domain, type, no_cache } = c.req.query();
	const cache = caches.default;
	let response = await cache.match(c.req.raw);
	if (response && !no_cache) {
		// If the response is cached and no_cache is not set, return the cached response
		const newHeaders = new Headers(response.headers);
		newHeaders.set("X-Worker-Cache", "HIT");
		return new Response(response.body, {
			status: response.status,
			statusText: response.statusText,
			headers: newHeaders,
		});
	}
	try {
		if (!domain || !type) {
			return c.json({ error: "Missing domain or type query parameters" }, 400);
		}
		const container = await getRandom(c.env.RESOLVER, 3);
		const containerResponse = await container.fetch(c.req.raw);
		const resp: LookupResponse = await containerResponse.json();
		const shortestTTL = getShortestTTL(resp);
		const isNoCache = no_cache === "true";
		const cacheControl = isNoCache
			? "no-cache"
			: typeof shortestTTL === "number" && shortestTTL > 0
				? `public, max-age=${shortestTTL}`
				: "no-cache";
		const response = new Response(JSON.stringify(resp), {
			headers: {
				"Content-Type": "application/json",
				"Cache-Control": cacheControl, // Use the calculated cache control header
			},
		});
		if (!no_cache) {
			c.executionCtx.waitUntil(cache.put(c.req.raw, response.clone()));
		}

		return response;
	} catch (err) {
		if (err instanceof Error) {
			console.error("Error fetch:", err.message);
			return new Response(err.message, { status: 500 });
		}
		console.error("Error fetch:", err);
		return new Response("Unknown error", { status: 500 });
	}
});

app.get("/health", async (c) => {
	const container = await getRandom(c.env.RESOLVER, 3);
	return container.fetch(c.req.raw);
});

app.get("/debug", async (c) => {
	const container = await getRandom(c.env.RESOLVER, 3);
	return container.fetch(c.req.raw);
});

app.get("/dns_types", async (c) => {
	const cache = caches.default;
	let response = await cache.match(c.req.raw);
	if (response) {
		const newHeaders = new Headers(response.headers);
		newHeaders.set("X-Worker-Cache", "HIT");
		return new Response(response.body, {
			status: response.status,
			statusText: response.statusText,
			headers: newHeaders,
		});
	}
	const container = await getRandom(c.env.RESOLVER, 3);
	const container_response = await container.fetch(c.req.raw);
	response = new Response(JSON.stringify(await container_response.json()), {
		headers: {
			"Cache-Control": `public, max-age=3600`, // Cache for 1 hour
			"X-Worker-Cache": "MISS",
			...container_response.headers,
		},
	});
	c.executionCtx.waitUntil(cache.put(c.req.raw, response.clone()));

	return response;
});

app.get("/dns_servers", async (c) => {
	const cache = caches.default;
	let response = await cache.match(c.req.raw);
	if (response) {
		const newHeaders = new Headers(response.headers);
		newHeaders.set("X-Worker-Cache", "HIT");
		return new Response(response.body, {
			status: response.status,
			statusText: response.statusText,
			headers: newHeaders,
		});
	}
	const container = await getRandom(c.env.RESOLVER, 3);
	const container_response = await container.fetch(c.req.raw);
	response = new Response(JSON.stringify(await container_response.json()), {
		headers: {
			"Content-Type": "application/json",
			"Cache-Control": `public, max-age=3600`, // Cache for 1 hour
			"X-Worker-Cache": "MISS",
			...container_response.headers,
		},
	});
	c.executionCtx.waitUntil(cache.put(c.req.raw, response.clone()));
	return response;
});

export function getShortestTTL(response: LookupResponse): number | null {
	if (!response.answers || response.answers.length === 0) {
		return null;
	}
	let minTTL: number | null = null;
	for (const answer of response.answers) {
		if (typeof answer.ttl === "number") {
			if (minTTL === null || answer.ttl < minTTL) {
				minTTL = answer.ttl;
			}
		}
	}
	return minTTL;
}

export default app;
