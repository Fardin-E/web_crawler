import { useState } from "react";


function UrlForm() {

    const [url, setUrl] = useState('');
    const [message] = useState('');

    const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
        event.preventDefault();

        try {
            const response = await fetch('http://localhost:8080/api/crawler/start', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ url: url }),
            });

            const data = await response.json();
            postMessage(data.message || 'Crawl started successfully!');
        } catch(error) {
            console.error('Failed to send URL: ', error);
            postMessage('Failed to start crawl. Check the console for errors');
        }
    };

    return (
        <form onSubmit={handleSubmit}>
            <input 
            type="text" 
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder="Enter URL to crawl"
            required
        />
        <button type="submit">Start Crawl</button>
        {message && <p>{message}</p>}
        </form>
    );
};

export default UrlForm;