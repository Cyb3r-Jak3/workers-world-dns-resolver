import { Container, getRandom } from "@cloudflare/containers";
import { Hono, type Context } from "hono";
import { fromHono, OpenAPIRoute, contentJson } from "chanfana";
import { version } from "../package.json";
import z from "zod";
export class RESOLVER extends Container {
	defaultPort = 8080;
	sleepAfter = "10m";
	enableInternet: boolean = true;
}

type Bindings = {
	RESOLVER: DurableObjectNamespace<RESOLVER>;
};

export type AppContext = Context<{ Bindings: Env }>;

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

const openapi = fromHono(app, {
	base: "/api/v1",
	schema: {
		info: {
			title: "DNS Resolver API",
			version: version,
			description: "API for DNS resolution using multiple resolvers.",
			contact: {
				name: "Cyb3rJak3",
				url: "https://cyberjake.xyz",
				email: "connect@cyberjake.xyz",
			},
		},
		servers: [
			{
				url: "https://world-dns-resolver.cyberjake.xyz/api/v1",
				description: "Production server",
			},
			{
				url: "http://localhost:8787/api/v1",
				description: "Local development server",
			},
		],
	},
});

// app.get("/lookup", async (c) => {
// 	const { domain, type, no_cache } = c.req.query();
// 	const cache = caches.default;
// 	let response = await cache.match(c.req.raw);
// 	if (response && !no_cache) {
// 		// If the response is cached and no_cache is not set, return the cached response
// 		const newHeaders = new Headers(response.headers);
// 		newHeaders.set("X-Worker-Cache", "HIT");
// 		return new Response(response.body, {
// 			status: response.status,
// 			statusText: response.statusText,
// 			headers: newHeaders,
// 		});
// 	}
// 	try {
// 		if (!domain || !type) {
// 			return c.json({ error: "Missing domain or type query parameters" }, 400);
// 		}
// 		const container = await getRandom(c.env.RESOLVER, 3);
// 		const containerResponse = await container.fetch(c.req.raw);
// 		const resp: LookupResponse = await containerResponse.json();
// 		const shortestTTL = getShortestTTL(resp);
// 		const isNoCache = no_cache === "true";
// 		const cacheControl = isNoCache
// 			? "no-cache"
// 			: typeof shortestTTL === "number" && shortestTTL > 0
// 				? `public, max-age=${shortestTTL}`
// 				: "no-cache";
// 		const response = new Response(JSON.stringify(resp), {
// 			headers: {
// 				"Content-Type": "application/json",
// 				"Cache-Control": cacheControl, // Use the calculated cache control header
// 			},
// 		});
// 		if (!no_cache) {
// 			c.executionCtx.waitUntil(cache.put(c.req.raw, response.clone()));
// 		}

// 		return response;
// 	} catch (err) {
// 		if (err instanceof Error) {
// 			console.error("Error fetch:", err.message);
// 			return new Response(err.message, { status: 500 });
// 		}
// 		console.error("Error fetch:", err);
// 		return new Response("Unknown error", { status: 500 });
// 	}
// });

class LookupEndpoint extends OpenAPIRoute {
	schema = {
		request: {
			query: z.object({
				domain: z.string().describe("The domain to look up"),
				type: z.string().describe("The DNS record type to look up, e.g., A, AAAA, CNAME, etc."),
				no_cache: z.string().optional().describe("If set to 'true', the response will not be cached"),
			}),
		},
		responses: {
			"200": {
				description: "Successful DNS lookup",
				...contentJson(
					z.object({
						question: z.string().describe("The domain being queried"),
						type: z.string().describe("The DNS record type queried"),
						answers: z
							.array(
								z.object({
									server: z.string().describe("The DNS server that provided the answer"),
									values: z.array(z.string()).describe("The resolved values for the domain"),
									server_address: z.string().describe("The address of the DNS server"),
									ttl: z.number().describe("Time to live for the DNS record in seconds"),
									duration: z.number().describe("Duration of the DNS query in nanoseconds"),
									duration_string: z.string().describe("Duration of the DNS query as a string"),
								}),
							)
							.describe("List of answers from different DNS servers"),
						location: z.string().describe("Geographical location of the DNS resolver"),
						region: z.string().describe("Region of the DNS resolver"),
						country: z.string().describe("Country of the DNS resolver"),
						total_duration: z.number().describe("Total duration of the DNS query in nanoseconds"),
						total_duration_string: z.string().describe("Total duration of the DNS query as a string"),
					}),
				),
			},
			"400": {
				description: "Bad Request - Missing domain or type query parameters",
				...contentJson(
					z.object({
						error: z.string().describe("Error message indicating missing parameters"),
					}),
				),
			},
		},
	};
	async handle(c: AppContext) {
		const data = await this.getValidatedData<typeof this.schema>();
		const queryParams = data.query;
		if (!queryParams) {
			return c.json({ error: "Missing query parameters" }, 400);
		}
		``;
		const { domain, type, no_cache } = queryParams;
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
	}
}

openapi.get("/lookup", LookupEndpoint);

class HealthEndpoint extends OpenAPIRoute {
	schema = {
		responses: {
			"200": {
				description: "Health check successful",
				...contentJson(
					z.object({
						status: z.string(),
					}),
				),
			},
			"503": {
				description: "Service Unavailable",
				...contentJson(
					z.object({
						status: z.string(),
					}),
				),
			},
		},
	};
	async handle(c: AppContext) {
		const container = await getRandom(c.env.RESOLVER, 3);
		return container.fetch(c.req.raw);
	}
}
openapi.get("/health", HealthEndpoint);

class DebugEndpoint extends OpenAPIRoute {
	schema = {
		responses: {
			"200": {
				description: "Health check successful",
				...contentJson(
					z.object({
						Version: z.string(),
						AppID: z.string(),
						Region: z.string(),
						Location: z.string(),
						Country: z.string(),
						DeploymentID: z.string(),
					}),
				),
			},
		},
	};
	async handle(c: AppContext) {
		const container = await getRandom(c.env.RESOLVER, 3);
		return container.fetch(c.req.raw);
	}
}

openapi.get("/debug", DebugEndpoint);

class DNSServersEndpoint extends OpenAPIRoute {
	schema = {
		responses: {
			"200": {
				description: "List of DNS servers",
				...contentJson(
					z.array(
						z.object({
							Name: z.string().describe("DNS server name"),
							Address: z.string().describe("DNS server address"),
							Port: z.number().describe("DNS server port"),
						}),
					),
				),
			},
		},
	};
	async handle(c: AppContext) {
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
	}
}
openapi.get("/dns_servers", DNSServersEndpoint);

class DNSTypesEndpoint extends OpenAPIRoute {
	schema = {
		responses: {
			"200": {
				description: "List of DNS types",
				...contentJson(z.array(z.string().describe("DNS type, e.g., A, AAAA, CNAME, etc."))),
			},
		},
	};
	async handle(c: AppContext) {
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
	}
}

openapi.get("/dns_types", DNSTypesEndpoint);

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
