import axios from "axios";
import { useNavigate } from "react-router-dom";


export const getApiBaseUrl = (): string => {
    return import.meta.env.VITE_API_URL;
};

const api = axios.create({
    withCredentials: true, // send/receive cookies (required for /oauth2/token and cookie-based auth)
});

const initializeApi = async () => {
    api.defaults.baseURL = getApiBaseUrl();
};

initializeApi();

api.interceptors.request.use((config) => {
    try{

        const baseUrl = getApiBaseUrl();
        config.baseURL = baseUrl;

        const username = localStorage.getItem("x-cf-uid");
        const deviceId = localStorage.getItem("x-cf-device-id");
        const bearer = localStorage.getItem("x-cf-bearer");
        const refresh = localStorage.getItem("x-cf-refresh");

        config.headers['x-cf-uid'] = username;
        config.headers['x-cf-device-id'] = deviceId;
        config.headers['x-cf-bearer'] = bearer;
        config.headers['x-cf-refresh'] = refresh;

        // config.withCredentials = true;

        return config;
    } catch(error) {
        console.error("Error in request interceptor:", error);
        return config;
    }
});

api.interceptors.response.use(
    (response) => response,
    async (error) => {
        const navigate = useNavigate();
        // Only attempt token refresh if we have a refresh token
        if (error.response?.status === 401 && 
            error.response?.data?.error && 
            (error.response.data.error.includes("unauthorized") || error.response.data.error.includes("empty"))) {
            try {
                const username = localStorage.getItem("x-cf-uid");
                
                if (!username) {
                    // Clear auth data and navigate to home if we don't have refresh token
                    localStorage.removeItem("x-cf-uid");
                    // Use setTimeout to ensure navigation happens after state updates
                    navigate("/login");
                    return Promise.reject(error);
                }

                try {
                    const baseUrl = await getApiBaseUrl();
                    const refreshResponse = await axios.post(`${baseUrl}/oauth2/token`, {
                        password: localStorage.getItem('x-cf-refresh'),
                        grant_type: 'refresh',
                        scope: "external",
                        username: username
                    }, {
                        headers: {
                            'Content-Type': 'application/json',
                            'x-cf-device-id': localStorage.getItem('x-cf-device-id'),
                            'x-cf-uid': username,
                            'x-cf-bearer': localStorage.getItem('x-cf-bearer'),
                            'x-cf-refresh': localStorage.getItem('x-cf-refresh'),
                        },
                        withCredentials: true, // cookies (x-cf-Refresh) must be sent
                    });

                    localStorage.setItem('x-cf-bearer', refreshResponse.data.access_token);
                    localStorage.setItem('x-cf-refresh', refreshResponse.data.refresh_token);

                    return api.request(error.config);
                } catch (refreshError) {
                    // Clear auth data and navigate to home on refresh failure
                    localStorage.removeItem("x-cf-uid");
                    // Use setTimeout to ensure navigation happens after state updates
                    setTimeout(() => {
                        navigate("/login");
                    }, 0);
                    return Promise.reject(refreshError);
                }
            } catch (error) {
                console.error('Error in 401 handling:', error);
                localStorage.removeItem("x-cf-uid");
                setTimeout(() => {
                    navigate("/login");
                }, 0);
                return Promise.reject(error);
            }
        }
        return Promise.reject(error);
    }
);

export default api;