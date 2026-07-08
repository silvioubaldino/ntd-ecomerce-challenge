import { Navigate, Route, Routes } from "react-router-dom";
import { Layout } from "./components/Layout";
import { ProductListPage } from "./features/products/ProductListPage";
import { ProductCreatePage } from "./features/products/ProductCreatePage";
import { ProductEditPage } from "./features/products/ProductEditPage";

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<Navigate to="/products" replace />} />
        <Route path="/products" element={<ProductListPage />} />
        <Route path="/products/new" element={<ProductCreatePage />} />
        <Route path="/products/:id/edit" element={<ProductEditPage />} />
      </Route>
    </Routes>
  );
}
