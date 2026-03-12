import { nav } from "framer-motion/client";
import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";

interface Company {
    name: string;
    ticker: string;
}

export default function Search() {
    const [query, setQuery] = useState('');
    const [results, setResults] = useState<Company[]>([]);
    const navigate = useNavigate();

    useEffect(() => {
        if (!query) return setResults([]);

        const timeout = setTimeout(async () => {
            const res = await fetch(`/v1/companies/search?q=${encodeURIComponent(query)}`);
            const data = await res.json();
            setResults(data);
        }, 300); // debounce

        return () => clearTimeout(timeout);
    }, [query]);

    const handleSelect = (ticker: string) => {
        setQuery('');
        navigate(`/client-app/stock?ticker=${ticker}`);
    };

    return (
        <div className="relative">
            <input
                className="border-b border-neutral-400 bg-transparent px-1 py-0.5 text-[10px] uppercase tracking-[0.18em] text-neutral-400 outline-none placeholder:text-neutral-400"
                type="text"
                placeholder="SEARCH"
                value={query}
                onChange={e => setQuery(e.target.value)}
            />
            {query && (
                <ul className="absolute top-full left-0 mt-1 w-48 bg-white border border-neutral-200 rounded shadow-md z-50">
                    {results.map(company => (
                        <li
                            key={company.ticker}
                            onClick={() => handleSelect(company.ticker)}
                            className="px-3 py-1.5 text-[10px] uppercase tracking-[0.18em] text-neutral-400 hover:bg-gray-100 cursor-pointer"
                        >
                            {company.ticker} — {company.name}
                        </li>
                    ))}
                </ul>
            )}
        </div>
    );
}