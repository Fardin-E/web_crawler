import React from 'react'
import type { CrawlResult } from '../types/types';

interface CrawlResultCardProps {
    crawlResult: CrawlResult;
}

const CrawlResultCard: React.FC<CrawlResultCardProps> = ({ crawlResult }) => {
    return (
        <div className="crawlResult-card">
            <p>URL: {crawlResult.url}</p>
            <p>Title: {crawlResult.info.title}</p>
        </div>
    );
};

export default CrawlResultCard;