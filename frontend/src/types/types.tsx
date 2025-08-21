// types.ts

export interface Token {
    name: string;
    value: string;
}

export interface Info {
    title: string;
    description: string;
    paragraphs: string[];
    links: Token[];
}

export interface CrawlResult {
    url: string;
    statusCode: number;
    contentType: string;
    responseTime: number;
    body: string;
    info: Info;
    isError: boolean;
}