import HeroSection from './Landing/components/HeroSection';
import FeaturesSection from './Landing/components/FeaturesSection';
import RolesSection from './Landing/components/RolesSection';
import ContactSection from './Landing/components/ContactSection';
import CallToActionSection from './Landing/components/CallToActionSection';
import FooterSection from './Landing/components/FooterSection';
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import './Landing/styles/Landing.css';

export const Landing = () => {
  return (
    <div className="landing-page">
      <HeroSection />
      <FeaturesSection />
      <RolesSection />
      <ContactSection />
      <CallToActionSection />
      <FooterSection />
      <ToastContainer
        position="top-right"
        autoClose={5000}
        hideProgressBar={false}
        newestOnTop
        closeOnClick
        rtl={false}
        pauseOnFocusLoss
        draggable
        pauseOnHover
        theme="light"
      />
    </div>
  );
};

export default Landing;
