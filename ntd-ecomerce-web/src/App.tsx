import { Navigate, Route, Routes } from "react-router-dom";
import { Layout } from "./components/Layout";
import { ProductListPage } from "./features/products/ProductListPage";
import { ProductCreatePage } from "./features/products/ProductCreatePage";
import { ProductEditPage } from "./features/products/ProductEditPage";
import { ProductImportPage } from "./features/products/ProductImportPage";
import { StoreSearchPage } from "./features/products/StoreSearchPage";
import { CartPage } from "./features/cart/CartPage";
import { CheckoutPage } from "./features/checkout/CheckoutPage";
import { OrderConfirmationPage } from "./features/checkout/OrderConfirmationPage";

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<Navigate to="/products" replace />} />
        <Route path="/products" element={<ProductListPage />} />
        <Route path="/products/new" element={<ProductCreatePage />} />
        <Route path="/products/import" element={<ProductImportPage />} />
        <Route path="/products/:id/edit" element={<ProductEditPage />} />
        <Route path="/store" element={<StoreSearchPage />} />
        <Route path="/cart" element={<CartPage />} />
        <Route path="/checkout" element={<CheckoutPage />} />
        <Route path="/orders/:orderId" element={<OrderConfirmationPage />} />
      </Route>
    </Routes>
  );
}
