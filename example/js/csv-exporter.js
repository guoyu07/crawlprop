const HCCrawler = require('headless-chrome-crawler');
const CSVExporter = require('headless-chrome-crawler/exporter/csv');

const FILE = './result.csv';

const exporter = new CSVExporter({
  file: FILE,
  fields: ['response.url', 'response.status', 'links.length'],
});

HCCrawler.launch({
  maxDepth: 2,
  exporter,
})
  .then(crawler => {
    crawler.queue('http://zero.webappsecurity.com');
    crawler.onIdle()
      .then(() => crawler.close());
  });