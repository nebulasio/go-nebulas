//! This file is included in paragroup.h
//! this file is only the part that defines for_each for general_iterator

template <class Iterator_t, class Functor_t>
static void for_each_impl_general_iterator(Iterator_t begin, Iterator_t end,
                                           Functor_t&& f, Entities_t& es,
                                           auto_partitioner* p) {
  // use a divide-and-conquer method to do for_each
  size_t divide_times = static_cast<int>(log2(ff::rt::concurrency()));
  uint64_t count = 0;
  Iterator_t t = begin;
  while (t != end) {
    t++;
    count++;
  }
  es = std::make_shared<::ff::internal::paras_with_lock>();
  for_each_impl_general_iterator_auto_partition(
      begin, end, std::forward<Functor_t>(f), es, count, divide_times);
}

template <class Iterator_t, class Functor_t>
static void for_each_impl_general_iterator_auto_partition(
    Iterator_t begin, Iterator_t end, Functor_t&& f, Entities_t& es,
    size_t count, size_t divide_times) {
  Iterator_t t = begin;
  Iterator_t bt = begin;
  size_t left = count;
  std::vector<para<void> > lgroup;
  while (divide_times != 0 && left != 1) {
    size_t sc = left / 2;
    left = left - sc;
    size_t c = 0;
    bt = t;
    while (c != sc) {
      t++;
      c++;
    }
    para<void> p;
    p([bt, t, sc, &f, &es, divide_times]() {
      for_each_impl_general_iterator_auto_partition(bt, t, std::move(f), es, sc,
                                                    divide_times - 1);
    });
    lgroup.push_back(p);
    divide_times--;
  }

  es->lock.lock();
  for (int i = 0; i < lgroup.size(); ++i) es->entities.push_back(lgroup[i]);
  es->lock.unlock();
  while (t != end) {
    f(t);
    t++;
  }
}

template <class Iterator_t, class Functor_t>
static void for_each_impl_general_iterator(Iterator_t begin, Iterator_t end,
                                           Functor_t&& f, Entities_t& es,
                                           simple_partitioner* p) {
  ff::thrd_id_t this_id = ff::rt::get_thrd_id();
  size_t concurrency = ff::rt::concurrency();  // TODO(A.A) this may be optimal.
  // TODO(A.A) we may have another partition approach!
  uint64_t count = 0;
  Iterator_t t = begin;
  while (t != end) {
    t++;
    count++;
  }
  uint64_t step = count / concurrency;
  uint64_t ls = count % concurrency;

  t = begin;
  es = std::make_shared<::ff::internal::paras_with_lock>();

  uint16_t counter = 0;  // added
  int32_t thrd_id = 0;
  while (t != end && thrd_id < concurrency) {
    if (thrd_id == this_id) {
      thrd_id++;
      continue;
    }
    Iterator_t tmp = t;
    count = 0;
    uint64_t upperbound = step;  // added
    if (counter < ls) upperbound++;
    counter++;
    while (tmp != end && count < upperbound) {
      tmp++;
      count++;
    }

    para<void> p;
    p([t, tmp, f]() {
      Iterator_t lt = t;
      while (lt != tmp) {
        f(lt);
        lt++;
      }
    });
    es->lock.lock();
    es->entities.push_back(p);
    es->lock.unlock();
    thrd_id++;
    t = tmp;
  }
  while (t != end) {
    f(t);
    t++;
  }
}
