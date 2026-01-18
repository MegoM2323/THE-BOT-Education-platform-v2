import HeroSection from './Landing/components/HeroSection';
import AboutSection from './Landing/components/AboutSection';
import LessonTypesSection from './Landing/components/LessonTypesSection';
import ExamplesSection from './Landing/components/ExamplesSection';
import SwapSystemSection from './Landing/components/SwapSystemSection';
import HowItWorksSection from './Landing/components/HowItWorksSection';
import ContactSection from './Landing/components/ContactSection';
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import './Landing/styles/Landing.css';

export const Landing = () => {
  return (
    <div className="landing-page">
      <HeroSection />
      <AboutSection />
      <LessonTypesSection />
      <ExamplesSection />
      <SwapSystemSection />
      <HowItWorksSection />
      <ContactSection />
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
