const fs = require('fs');
const commentJson = require('comment-json');

const file = 'wrangler.jsonc';
const accountId = process.env.CLOUDFLARE_ACCOUNT_ID;

// Grab the tag from the first CLI arg, default to "latest" if not provided
const tag = process.argv[2] || 'latest';

const image = `registry.cloudflare.com/${accountId}/workers-world-dns-resolver:${tag}`;

const data = commentJson.parse(fs.readFileSync(file, 'utf8'));
if (data.env && data.env.prd && Array.isArray(data.env.prd.containers)) {
  data.env.prd.containers.forEach(c => { c.image = image; });
  fs.writeFileSync(file, commentJson.stringify(data, null, 2));
} else {
  console.error('Could not find env.prd.containers in wrangler.jsonc');
  process.exit(1);
}