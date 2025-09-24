import { useState } from "react";


function ControlPanel() {

    const [url, setUrl] = useState('');
    
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
        <div className="flex flex-col min-h-screen items-start p-4">
            <div className="flex w-full text-white">
                <h2 className="text-xl font-bold">Control Panel</h2>
            </div>
            <div className="flex my-4">
                <form onSubmit={handleSubmit}>
                    <input
                    type="text"
                    value={url}
                    onChange={(e) => setUrl(e.target.value)}
                    placeholder="Enter URL to crawl"
                    required
                    />
                </form>
            </div>
            <div className="flex space-x-4">
                <button type="submit">Start Crawl</button>
                <button type="submit">Reset</button>
            </div>
            
        </div>
    );
};

export default ControlPanel;