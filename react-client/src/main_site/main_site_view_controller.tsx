import { useEffect } from 'react';
import Home from './home';
import MainSiteNavbar from './components/navbar';
import AuthLogin from './auth_login';
import AuthCreate from './auth_create';
import AuthVerify from './auth_verify';

const MainSiteViewController = () => {
    const currentTab = window.location.pathname.split("/")[1];

    useEffect(() => {
        const pageNames: Record<string, string> = {
            home: 'Home',
            login: 'Log in',
            signup: 'Sign up',
            verify: 'Verify email',
        };
        const pageName = pageNames[currentTab] ?? 'Home';
        document.title = `CacheFlow | ${pageName}`;
    }, [currentTab]);

    let component = null;

    switch(currentTab) {
        case "profile":
            component = <Home />;
            break;

        case "login":
            component = <AuthLogin />;
            break;  
        case "signup":
            component = <AuthCreate />;
            break;
        case "verify":
            component = <AuthVerify />;
            break;
        default:
            component = <Home />;
            break;
    }

  return (
    <div className="relative flex flex-col w-full h-screen bg-black">

        {
            currentTab !== "login" && currentTab !== "signup" && (
                <div className="fixed top-0 left-0 w-full z-10">
                    <MainSiteNavbar />
                </div>
            )
        }
        
        
        {component}
    </div>
  )
}

export default MainSiteViewController