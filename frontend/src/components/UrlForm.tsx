import React, { useState } from "react";

export interface StartCrawlRequest {
    url: string;
}

function Form() {
    const [url, setUrl] = useState<string>('');

    const handleCrawlStart = async () => {
        // No need for event.preventDefault() here
        if (!url) {
            alert('Please enter a URL.');
            return;
        }

        const requestBody: StartCrawlRequest = { url: url };

        try {
            const response = await fetch('http://localhost:8080/api/crawler/start', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(requestBody),
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status ${response.status}`);
            }

            const data = await response.json();
            console.log('Crawler started successfully: ', data);
            
        } catch (error) {
            console.error('Failed to start crawler: ', error);
        }
    };

    const handleUrlChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setUrl(event.target.value);
    };

    return (
        <div className="form-container">
            <label className="form-label">
                Enter Url Here:
                <input
                    type="url"
                    className="form-input"
                    value={url}
                    onChange={handleUrlChange}
                />
            </label>
            <button type="button" className="form-button" onClick={handleCrawlStart}>
                Send
            </button>
        </div>
    );
}

export default Form;