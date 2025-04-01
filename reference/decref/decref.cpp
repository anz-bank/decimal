#include "decref.h"

#include <stdio.h>
#include <string.h>

#include <boost/decimal.hpp>
#include <iostream>
#include <sstream>

using namespace boost::decimal;

// Inline function to convert decimal64 to Dec64
inline Dec64 f64(decimal64 d) {
  return *(Dec64*)&d;
}

// Inline function to convert Dec64 to decimal64
inline decimal64 t64(Dec64 d) {
  return *(decimal64*)(&d);
}

Dec64 parse64(const char* str) {
  decimal64 d;
  std::istringstream(str) >> d;
  return f64(d);
}

char* string64(Dec64 wrapper) {
  std::ostringstream oss;
  oss << std::scientific
      << std::setprecision(std::numeric_limits<double>::max_digits10)
      << t64(wrapper) << std::flush;
  return strdup(oss.str().c_str());
}

Dec64 frombid64(uint64_t b) { return f64(from_bid<decimal64>(b)); }
uint64_t tobid64(Dec64 a) { return to_bid(t64(a)); }

int isNaN64(Dec64 a) { return isnan(t64(a)); }

Dec64 add64(Dec64 a, Dec64 b) { return f64(t64(a) + t64(b)); }
Dec64 sub64(Dec64 a, Dec64 b) { return f64(t64(a) - t64(b)); }
Dec64 mul64(Dec64 a, Dec64 b) { return f64(t64(a) * t64(b)); }
Dec64 quo64(Dec64 a, Dec64 b) { return f64(t64(a) / t64(b)); }
